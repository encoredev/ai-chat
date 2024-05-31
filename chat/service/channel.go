package chat

import (
	"context"

	"github.com/cockroachdb/errors"

	"encore.app/bot"
	botdb "encore.app/bot/db"
	"encore.app/chat/service/client"
	"encore.app/chat/service/db"
	"encore.app/llm/service"
	"encore.dev/rlog"
	"encore.dev/types/uuid"
)

type ListChannelsResponse struct {
	Channels []*db.Channel
}

// GetChannel returns channel information by ID.
//
//encore:api private method=GET path=/chat/channels/:id
func (svc *Service) GetChannel(ctx context.Context, id uuid.UUID) (*db.Channel, error) {
	queries := db.New()
	channel, err := queries.GetChannel(ctx, chatdb.Stdlib(), id)
	if err != nil {
		return nil, errors.Wrap(err, "get channel")
	}
	return channel, nil

}

type ListChannelResponse struct {
	Channels []*db.Channel
}

// ListChannels returns a list of all channels from all providers.
//
//encore:api public method=GET path=/chat/channels
func (svc *Service) ListChannels(ctx context.Context) (*ListChannelsResponse, error) {
	queries := db.New()
	channels, err := queries.ListChannels(ctx, chatdb.Stdlib())
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}
	return &ListChannelsResponse{Channels: channels}, nil
}

// AddBotToChannel adds a bot to a channel. It will trigger a join event in the chat provider.
// It will also create a introduction task in the LLM service which will be sent to the channel.
//
//encore:api public method=POST path=/chat/channels/:channelID/bots/:botID
func (svc *Service) AddBotToChannel(ctx context.Context, channelID uuid.UUID, botID uuid.UUID) error {
	q := db.New()
	c, err := q.GetChannel(ctx, chatdb.Stdlib(), channelID)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}
	_, err = q.UpsertBotChannel(ctx, chatdb.Stdlib(), db.UpsertBotChannelParams{
		Bot:      botID,
		Channel:  channelID,
		Provider: c.Provider,
	})
	if err != nil {
		return errors.Wrap(err, "upsert botID channel")
	}
	b, err := bot.Get(ctx, botID)
	if err != nil {
		return errors.Wrap(err, "get bot")
	}
	_, err = svc.loadChannelHistory(ctx, c)
	if err != nil {
		return errors.Wrap(err, "load channel history")
	}

	prov, ok := svc.providers[c.Provider]
	if !ok {
		return errors.Newf("unknown provider %v", c.Provider)
	}
	err = prov.GetChannelClient(ctx, c.ProviderID).Join(ctx, b)
	if err != nil {
		return errors.Wrap(err, "join channel")
	}
	err = svc.publishLLMTasks(ctx, llm.TaskTypeJoin, []*botdb.Bot{b}, c, "")
	if err != nil {
		return errors.Wrap(err, "publish chat event")
	}
	return nil
}

// RemoveBotFromChannel removes a bot from a channel. It will trigger a leave event in the chat provider.
// It will also create a leave task in the LLM service which will be sent to the channel.
//
//encore:api public method=DELETE path=/chat/channels/:channelID/bots/:botID
func (svc *Service) RemoveBotFromChannel(ctx context.Context, channelID uuid.UUID, botID uuid.UUID) error {
	queries := db.New()
	_, err := queries.RemoveBotChannel(ctx, chatdb.Stdlib(), db.RemoveBotChannelParams{
		Bot:     botID,
		Channel: channelID,
	})
	if err != nil {
		return errors.Wrap(err, "remove bot channel")
	}
	channel, err := svc.GetChannel(ctx, channelID)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}
	bot, err := bot.Get(ctx, botID)
	if err != nil {
		return errors.Wrap(err, "get bot")
	}

	err = svc.publishLLMTasks(ctx, llm.TaskTypeLeave, []*botdb.Bot{bot}, channel, "")
	if err != nil {
		return errors.Wrap(err, "publish chat event")
	}
	return nil
}

// initChannels loads all channels from all providers and inserts them into the database.
// It will also load the history of all channels with bots.
func (svc *Service) initChannels(ctx context.Context) error {
	for typ, provider := range svc.providers {
		channels, err := provider.ListChannels(ctx)
		if err != nil {
			rlog.Warn("list channels failed", "provider", typ, "err", err)
			continue
		}
		for _, channel := range channels {
			_, err := svc.insertChannel(ctx, channel)
			if err != nil {
				return errors.Wrap(err, "insert channel")
			}
		}
	}
	botChannels, err := db.New().ListChannelsWithBots(ctx, chatdb.Stdlib())
	if err != nil {
		return errors.Wrap(err, "list channels with bots")
	}
	for _, pc := range botChannels {
		if _, err := svc.loadChannelHistory(ctx, pc); err != nil {
			rlog.Warn("load channel history failed", "channel", pc.ID, "err", err)
		}
	}
	return nil
}

// insertChannel inserts a channel into the database. It's triggered when a new channel is discovered from
// a chat provider.
func (svc *Service) insertChannel(ctx context.Context, channel client.ChannelInfo) (*db.Channel, error) {
	queries := db.New()
	dbChannel, err := queries.UpsertChannel(ctx, chatdb.Stdlib(), db.UpsertChannelParams{
		ProviderID: channel.ID,
		Provider:   channel.Provider,
		Name:       channel.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "upsert channel")
	}
	return dbChannel, nil
}
