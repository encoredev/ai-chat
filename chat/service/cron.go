package chat

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"

	"encore.app/chat/service/db"
	"encore.dev/config"
	"encore.dev/cron"
	"encore.dev/rlog"
)

type Config struct {
	InitConversationIntervalMinutes config.Int
}

var cfg = config.Load[*Config]()

// This cron job initiates conversations with users in channels that have not had a conversation in a while
var _ = cron.NewJob("initiate-convos", cron.JobConfig{
	Title:    "Initiate Conversations",
	Every:    1 * cron.Hour,
	Endpoint: InitiateConversation,
})

// InitiateConversation initiates conversations with users in channels that have not had a conversation in a while
//
//encore:api private
func (svc *Service) InitiateConversation(ctx context.Context) error {
	rlog.Debug("Initiating conversations")
	q := db.New()
	channels, err := q.ListChannelsWithBots(ctx, chatdb.Stdlib())
	if err != nil {
		return errors.Wrap(err, "list channels with bots")
	}
	for _, channel := range channels {
		latest, err := q.LatestBotMessageInChannel(ctx, chatdb.Stdlib(), channel.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "latest message in channel")
		}
		if time.Since(latest.Timestamp) < time.Duration(cfg.InitConversationIntervalMinutes())*time.Minute {
			continue
		}
		bots, err := q.ListBotsInChannel(ctx, chatdb.Stdlib(), channel.ID)
		if err != nil {
			return errors.Wrap(err, "list bots in channel")
		}
		err = svc.InstructBot(ctx, &InstructRequest{
			Bots:        bots,
			Channel:     channel.ID,
			Instruction: "Continue a discussion with a character or start a completely random new one",
		})
		if err != nil {
			return errors.Wrap(err, "instruct bot")
		}
	}
	return nil
}
