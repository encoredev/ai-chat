package encorechat

import (
	"context"
	"embed"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	_ "github.com/gorilla/websocket"

	"encore.app/chat/provider"
	"encore.app/chat/provider/encorechat/chat"
	"encore.app/chat/service/client"
	chatdb "encore.app/chat/service/db"
	"encore.dev/config"
	"encore.dev/rlog"
	"encore.dev/types/uuid"
)

type Config struct {
	Enabled config.Bool
}

var cfg = config.Load[*Config]()

//encore:service
type Service struct {
	hub  *chat.Hub
	data *DataSource
}

func initService() (*Service, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	svc := &Service{}
	svc.hub = chat.NewHub(context.Background(), func(ctx context.Context, msg *chat.ClientMessage) error {
		return svc.SendMessage(ctx, msg.ConversationId, &provider.SendMessageRequest{
			Content: string(msg.Content),
			UserID:  msg.UserId,
			Type:    msg.Type,
		})
	})
	return svc, nil
}

// Ping returns an error if the Discord service is not available.
// encore:api private
func (p *Service) Ping(ctx context.Context) error {
	if p == nil {
		return errors.New("Encore chat service is not available for this environment. Set Enabled to true in the config to enable it.")
	}
	return nil
}

//go:embed static/build/*
var staticFiles embed.FS

//encore:api public raw path=/!fallback
func (s *Service) ServeHTML(w http.ResponseWriter, r *http.Request) {
	if !cfg.Enabled() {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	path = "/static/build" + path
	http.ServeFileFS(w, r, staticFiles, path)
}

//encore:api public raw path=/encorechat/subscribe
func (s *Service) Subscribe(w http.ResponseWriter, r *http.Request) {
	if !cfg.Enabled() {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}
	err := s.hub.Subscribe(w, r)
	if err != nil {
		rlog.Error("upgrade", "error", err)
		http.Error(w, "upgrade connection", http.StatusInternalServerError)
		return
	}
}

//encore:api private method=POST path=/encorechat/channels/:channelID/messages
func (s *Service) SendMessage(ctx context.Context, channelID string, req *provider.SendMessageRequest) error {
	switch req.Type {
	case "join":
		channel, created, err := s.data.UpsertChannel(ctx, channelID)
		if err != nil {
			return errors.Wrap(err, "upsert channel")
		}
		if created {
			bots, err := s.data.RandomizeBots(ctx, 3)
			if err != nil {
				return errors.Wrap(err, "randomize bots")
			}

		}
	}

	author := client.User{
		ID:   req.UserID,
		Name: req.UserID,
	}
	if req.Bot != nil {
		author.ID = req.Bot.ID.String()
		author.Name = req.Bot.Name
		author.BotID = req.Bot.ID
	}
	s.hub.BroadCast(ctx, &chat.ClientMessage{
		Type:           req.Type,
		UserId:         req.UserID,
		ConversationId: channelID,
		Content:        req.Content,
	})
	_, err := provider.InboxTopic.Publish(ctx, &client.Message{
		Provider:   chatdb.ProviderEncorechat,
		ProviderID: uuid.Must(uuid.NewV4()).String(),
		ChannelID:  channelID,
		Author:     author,
		Content:    req.Content,
		Time:       time.Now(),
		Type:       req.Type,
	})
	return errors.Wrap(err, "publish message")
}
