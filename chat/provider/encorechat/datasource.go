package encorechat

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"

	"encore.app/bot"
	botdb "encore.app/bot/db"
	"encore.app/chat/provider"
	"encore.app/chat/service/client"
	chatdb "encore.app/chat/service/db"
	"encore.app/pkg/fns"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

var db = sqldb.Named("chat")

type DataSource struct{}

func (d *DataSource) RandomizeBots(ctx context.Context, count int) ([]*botdb.Bot, error) {
	resp, err := bot.List(ctx, &bot.ListBotRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "list bots")
	}
	return fns.SelectRandom(resp.Bots, count), nil
}

func (d *DataSource) UpsertChannel(ctx context.Context, id string) (*chatdb.Channel, error) {
	q := chatdb.New()
	channel, err := q.GetChannelByProviderID(ctx, db.Stdlib(), chatdb.GetChannelByProviderIDParams{
		ProviderID: id,
		Provider:   chatdb.ProviderEncorechat,
	})
	if errors.Is(err, sql.ErrNoRows) {
		channel, err = q.UpsertChannel(ctx, db.Stdlib(), chatdb.UpsertChannelParams{
			ProviderID: id,
			Provider:   chatdb.ProviderEncorechat,
			Name:       id,
		})
		bots, err := d.RandomizeBots(ctx, 3)
		if err != nil {
			return nil, errors.Wrap(err, "randomize bots")
		}
		_, err = provider.InboxTopic.Publish(ctx, &client.Message{
			Provider:   chatdb.ProviderEncorechat,
			ProviderID: uuid.Must(uuid.NewV4()).String(),
			ChannelID:  id,
			Time:       time.Now(),
			Type:       "channel_created",
		})
		return channel, errors.Wrap(err, "upsert channel")
	} else if err != nil {
		return nil, errors.Wrap(err, "get channel")
	}
	return channel, nil
}

func (d *DataSource) GetChannelUsers(ctx context.Context, c *chatdb.Channel) ([]*chatdb.User, []*botdb.Bot, error) {
	q := chatdb.New()
	members, err := q.ListUsersInChannel(ctx, db.Stdlib(), c.ID)
	botsIds := fns.Map(members, func(m *chatdb.User) uuid.UUID { return m.ID })
	bots, err := bot.List(ctx, &bot.ListBotRequest{IDs: botsIds})
	if err != nil {
		return nil, nil, errors.Wrap(err, "list bots")
	}
	return members, bots.Bots, nil
}

func (d *DataSource) GetChannelMessage(ctx context.Context, c *chatdb.Channel) ([]*chatdb.Message, error) {
	q := chatdb.New()
	messages, err := q.ListMessagesInChannel(ctx, db.Stdlib(), c.ID)
	return messages, errors.Wrap(err, "list messages")
}
