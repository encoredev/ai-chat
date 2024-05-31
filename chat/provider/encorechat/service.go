package encorechat

import (
	"context"
	"embed"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/websocket"

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
	upgrader websocket.Upgrader

	hub *chat.Hub

	// Service context
	ctx context.Context
}

func initService() (*Service, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	svc := &Service{}
	svc.hub = chat.NewHub(context.Background(), func(ctx context.Context, channel, author string, content []byte) error {
		return svc.SendMessage(ctx, channel, &provider.SendMessageRequest{
			Content: string(content),
			UserID:  author,
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

//go:embed static/*
var staticFiles embed.FS

//encore:api public raw path=/encorechat/demo/*path
func (s *Service) ServeHTML(w http.ResponseWriter, r *http.Request) {
	if !cfg.Enabled() {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/encorechat/demo")
	if path == "/" {
		path = "/index.html"
	}
	path = "/static" + path
	http.ServeFileFS(w, r, staticFiles, path)
}

//encore:api public raw path=/encorechat/channels/:channelID/subscribe/:user
func (s *Service) Subscribe(w http.ResponseWriter, r *http.Request) {
	if !cfg.Enabled() {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/encorechat/channels/")
	channelID, userID, ok := strings.Cut(path, "/subscribe/")
	if !ok {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	err := s.hub.Subscribe(channelID, userID, w, r)
	if err != nil {
		rlog.Error("upgrade", "error", err)
		http.Error(w, "upgrade connection", http.StatusInternalServerError)
		return
	}
}

//encore:api private method=POST path=/encorechat/channels/:channelID/messages
func (s *Service) SendMessage(ctx context.Context, channelID string, req *provider.SendMessageRequest) error {
	author := client.User{
		ID:   req.UserID,
		Name: req.UserID,
	}
	if req.Bot != nil {
		author.ID = req.Bot.ID.String()
		author.Name = req.Bot.Name
		author.BotID = req.Bot.ID
	}
	s.hub.BroadCast(ctx, channelID, author.Name, req.Content)
	_, err := provider.MessageTopic.Publish(ctx, &client.Message{
		Provider:   chatdb.ProviderEncorechat,
		ProviderID: uuid.Must(uuid.NewV4()).String(),
		ChannelID:  channelID,
		Author:     author,
		Content:    req.Content,
		Time:       time.Now(),
	})
	return errors.Wrap(err, "publish message")
}
