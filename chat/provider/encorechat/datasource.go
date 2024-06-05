package encorechat

import (
	"context"
	"slices"

	"github.com/cockroachdb/errors"

	"encore.app/bot"
	botdb "encore.app/bot/db"
	"encore.app/chat/service/db"
	"encore.app/pkg/fns"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

var chatDb = sqldb.Named("chat")

type DataSource struct{}

func (d *DataSource) GetChannel(ctx context.Context, id string) (*db.Channel, bool) {
	q := db.New()
	channel, err := q.GetChannelByProviderID(ctx, chatDb.Stdlib(), db.GetChannelByProviderIDParams{
		ProviderID: id,
		Provider:   db.ProviderEncorechat,
	})
	if err != nil {
		return nil, false
	}
	return channel, true
}

func (d *DataSource) GetChannelUsers(ctx context.Context, c *db.Channel) ([]*db.User, []*botdb.Bot, error) {
	q := db.New()
	members, err := q.ListUsersInChannel(ctx, chatDb.Stdlib(), c.ID)
	botUsers := fns.Filter(members, func(m *db.User) bool { return m.BotID != nil })
	botsIds := fns.Map(botUsers, func(m *db.User) uuid.UUID { return *m.BotID })
	bots, err := bot.List(ctx, &bot.ListBotRequest{IDs: botsIds})
	if err != nil {
		return nil, nil, errors.Wrap(err, "list bots")
	}
	return members, bots.Bots, nil
}

func (d *DataSource) GetChannelMessages(ctx context.Context, c *db.Channel) ([]*db.Message, error) {
	q := db.New()
	messages, err := q.ListMessagesInChannel(ctx, chatDb.Stdlib(), c.ID)
	if err != nil {
		return nil, errors.Wrap(err, "list messages")
	}
	slices.Reverse(messages)
	return messages, nil
}