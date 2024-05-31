package chat

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/cockroachdb/errors"

	"encore.app/bot"
	botdb "encore.app/bot/db"
	"encore.app/chat/service/clients"
	"encore.app/chat/service/db"
	llmprovider "encore.app/llm/provider"
	"encore.app/llm/service"
	fns "encore.app/pkg/fns"
	"encore.dev/rlog"
	"encore.dev/types/uuid"
)

// getChannelHistory returns the message history for a channel. It doesn't fetch the messages from the provider,
// but rather from the database.
func (svc *Service) getChannelHistory(ctx context.Context, channelID db.ChannelID) ([]*db.Message, error) {
	queries := db.New()
	msgs, err := queries.ListMessagesInChannel(ctx, chatdb.Stdlib(), channelID)
	if err != nil {
		return nil, errors.Wrap(err, "list messages by channel")
	}
	// We fetch the messages in reverse order to get the latest messages, so we need to reverse them back
	slices.Reverse(msgs)
	return msgs, nil
}

// getChannelUsers returns the users in a channel. It always includes the admin user which is used to instruct the bot
// when sending commands to the LLM providers
func (svc *Service) getChannelUsers(ctx context.Context, channelID db.ChannelID) ([]*db.User, error) {
	queries := db.New()
	users, err := queries.ListUsersInChannel(ctx, chatdb.Stdlib(), channelID)
	if err != nil {
		return nil, errors.Wrap(err, "list users by channel")
	}
	// Admin has to be included to ensure the bot can be instructed
	users = append(users, db.Admin)
	return users, nil
}

type InstructRequest struct {
	Bots        []db.BotID
	Channel     db.ChannelID
	Instruction string
}

// InstructBot sends an instruction to a set of bots in a channel. It publishes a task to the LLM provider which
// decides how to handle the instruction. It can be used to e.g. manually trigger a message from a bot.
// Instructions are not sent to the chat providers
//
//encore:api public method=POST path=/discord/instruct
func (svc *Service) InstructBot(ctx context.Context, req *InstructRequest) error {
	q := db.New()
	_, err := q.InsertMessage(ctx, chatdb.Stdlib(), db.InsertMessageParams{
		ChannelID: req.Channel,
		AuthorID:  db.Admin.ID,
		Content:   req.Instruction,
		Timestamp: time.Now().UTC(),
	})
	if err != nil {
		return errors.Wrap(err, "insert message")
	}
	bots, err := bot.List(ctx, &bot.ListBotRequest{IDs: req.Bots})
	if err != nil {
		return errors.Wrap(err, "get bot")
	}
	channel, err := svc.GetChannel(ctx, req.Channel)
	err = svc.publishLLMTasks(ctx, llm.TaskTypeInstruct, bots.Bots, channel, req.Instruction)
	return errors.Wrap(err, "publish llm task")

}

func (svc *Service) publishLLMTasks(ctx context.Context, typ llm.TaskType, bots []*botdb.Bot, channel *db.Channel, adminPrompt string) error {
	msgs, err := svc.getChannelHistory(ctx, channel.ID)
	if err != nil {
		return errors.Wrap(err, "get channel history")
	}
	users, err := svc.getChannelUsers(ctx, channel.ID)
	if err != nil {
		return errors.Wrap(err, "get channel users")
	}

	botsByProvider := make(map[string][]*botdb.Bot)
	for _, b := range bots {
		botsByProvider[b.Provider] = append(botsByProvider[b.Provider], b)
	}
	for prov, bots := range botsByProvider {
		_, err := llm.TaskTopic.Publish(ctx, &llm.Task{
			Type: typ,
			Request: &llmprovider.ChatRequest{
				Bots:      bots,
				Users:     users,
				Channel:   channel,
				Messages:  msgs,
				SystemMsg: adminPrompt,
				Provider:  prov,
			},
		})
		if err != nil {
			rlog.Warn("publish llm task", "error", err)
		}
	}
	return nil
}

func (svc *Service) ProcessBotMessage(ctx context.Context, event *llm.BotResponse) error {
	prov, ok := svc.providers[event.Channel.Provider]
	if !ok {
		return errors.New("provider not found")
	}
	pc := prov.GetChannel(ctx, event.Channel.ProviderID)
	for _, msg := range event.Messages {
		err := pc.Send(ctx, msg.Bot, msg.Content)
		if err != nil {
			rlog.Warn("send message", "error", err)
		}
	}
	bots := make(map[*botdb.Bot]struct{})
	for _, msg := range event.Messages {
		bots[msg.Bot] = struct{}{}
	}
	if event.TaskType == llm.TaskTypeLeave {
		for b, _ := range bots {
			err := prov.GetChannel(ctx, event.Channel.ProviderID).Leave(ctx, b)
			if err != nil {
				return errors.Wrap(err, "leave channel")
			}
		}
	}
	return nil
}

//encore:api private path=/message method=POST
func (svc *Service) ProcessInboundMessage(ctx context.Context, msg *client.Message) error {
	msgs, err := svc.handleProviderMessages(ctx, msg.Provider, msg)
	if err != nil {
		return errors.Wrap(err, "handle provider messages")
	}
	if len(msgs) == 0 {
		return nil
	}
	q := db.New()
	author, err := q.GetUser(ctx, chatdb.Stdlib(), msgs[0].AuthorID)
	if err != nil {
		return errors.Wrap(err, "get user")
	}
	if author.BotID != nil {
		return nil
	}
	botIDs, err := q.ListBotsInChannel(ctx, chatdb.Stdlib(), msgs[0].ChannelID)
	if err != nil {
		return errors.Wrap(err, "list bots in channel")
	}
	if len(botIDs) == 0 {
		return nil
	}
	bots, err := bot.List(ctx, &bot.ListBotRequest{IDs: botIDs})
	if err != nil {
		return errors.Wrap(err, "list bots")
	}
	channel, err := svc.GetChannel(ctx, msgs[0].ChannelID)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}
	err = svc.publishLLMTasks(ctx, llm.TaskTypeContinue, bots.Bots, channel, "")
	return errors.Wrap(err, "publish llm task")
}

func (svc *Service) loadChannelHistory(ctx context.Context, channel *db.Channel) ([]*db.Message, error) {
	queries := db.New()
	msg, err := queries.LatestMessageInChannel(ctx, chatdb.Stdlib(), channel.ID)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return nil, errors.Wrap(err, "latest message in channel")
	}
	prov, ok := svc.providers[channel.Provider]
	if !ok {
		return nil, errors.New("provider not found")
	}
	messages, err := prov.GetChannel(ctx, channel.ProviderID).ListMessages(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "list messages")
	}
	return svc.handleProviderMessages(ctx, channel.Provider, messages...)
}

func (svc *Service) handleProviderMessages(ctx context.Context, providerName db.Provider, messages ...*client.Message) ([]*db.Message, error) {
	q := db.New()
	prov, ok := svc.providers[providerName]
	if !ok {
		return nil, errors.New("provider not found")
	}
	users, err := q.ListUsersByProvider(ctx, chatdb.Stdlib(), providerName)
	if err != nil {
		return nil, errors.Wrap(err, "list users by provider")
	}
	userByID := fns.ToMap(users, func(u *db.User) client.UserID { return u.ProviderID })
	channels, err := q.ListChannelsByProvider(ctx, chatdb.Stdlib(), providerName)
	if err != nil {
		return nil, errors.Wrap(err, "list channels by provider")
	}
	channelByID := fns.ToMap(channels, func(c *db.Channel) client.ChannelID { return c.ProviderID })

	var insertedMessages []*db.Message
	publishedChannels := make(map[db.ChannelID]struct{})
	for _, msg := range messages {
		author, ok := userByID[msg.Author.ID]
		if !ok {
			params := db.InsertUserParams{
				Provider:   providerName,
				ProviderID: msg.Author.ID,
				Name:       msg.Author.Name,
			}
			if msg.Author.BotID != uuid.Nil {
				params.BotID = &msg.Author.BotID
			} else {
				user, err := prov.GetUser(ctx, msg.Author.ID)
				if err != nil {
					return nil, errors.Wrap(err, "get user")
				}
				if user != nil {
					params.Profile = user.Profile
					params.Name = user.Name
				}
			}
			author, err = q.InsertUser(ctx, chatdb.Stdlib(), params)
			if err != nil {
				return nil, errors.Wrap(err, "insert user")
			}
			userByID[msg.Author.ID] = author
		}
		channel, ok := channelByID[msg.Channel]
		if !ok {
			cInfo, err := prov.GetChannel(ctx, msg.Channel).Info(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "get channel info")
			}
			channel, err = svc.insertChannel(ctx, cInfo)
			if err != nil {
				return nil, errors.Wrap(err, "insert channel")
			}
		}
		publishedChannels[channel.ID] = struct{}{}
		dbMsg, err := q.InsertMessage(ctx, chatdb.Stdlib(), db.InsertMessageParams{
			ChannelID:  channel.ID,
			AuthorID:   author.ID,
			Content:    msg.Content,
			Timestamp:  msg.Time,
			ProviderID: msg.ProviderID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			continue
		} else if err != nil {
			return nil, errors.Wrap(err, "insert message")
		}
		insertedMessages = append(insertedMessages, dbMsg)
	}
	return insertedMessages, nil
}
