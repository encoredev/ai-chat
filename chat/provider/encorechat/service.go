package encorechat

import (
	"context"
	"embed"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	_ "github.com/gorilla/websocket"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider"
	"encore.app/chat/provider/encorechat/chat"
	chatdb "encore.app/chat/service/db"
	"encore.app/pkg/fns"
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
	svc.hub = chat.NewHub(context.Background(), svc.handleClientMessage)
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

//encore:api public raw path=/encorechat/demo/!fallback
func (s *Service) ServeHTML(w http.ResponseWriter, r *http.Request) {
	if !cfg.Enabled() {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/encorechat/demo")
	if path == "/" || strings.HasPrefix(path, "/chat") {
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

//encore:api public method=POST path=/encorechat/channels/:channelID/join
func (s *Service) JoinChannel(ctx context.Context, channelID string, bot *botdb.Bot) error {
	s.hub.BroadCast(ctx, &chat.ClientMessage{
		Type:           "join",
		UserId:         "b-" + bot.ID.String(),
		ConversationId: channelID,
		Content:        bot.Profile,
		Avatar:         bot.GetAvatarURL(),
		Username:       bot.Name,
	})
	return nil
}

func (s *Service) sendChannelHistory(ctx context.Context, channelID string, afterID string, client *chat.Client) error {
	channel, ok := s.data.GetChannel(ctx, channelID)
	if !ok {
		_, err := provider.InboxTopic.Publish(ctx, &provider.Message{
			Provider:   chatdb.ProviderEncorechat,
			ProviderID: uuid.Must(uuid.NewV4()).String(),
			ChannelID:  channelID,
			Time:       time.Now(),
			Type:       "channel_created",
		})
		return errors.Wrap(err, "publish message")
	} else {
		users, bots, err := s.data.GetChannelUsers(ctx, channel)
		if err != nil {
			return errors.Wrap(err, "get channel users")
		}
		for _, user := range users {
			if user.BotID == nil {
				client.SendMessage(&chat.ClientMessage{
					Type:           "join",
					UserId:         user.Name,
					ConversationId: channelID,
					Username:       user.Name,
				})
			}
		}
		for _, bot := range bots {
			client.SendMessage(&chat.ClientMessage{
				Type:           "join",
				UserId:         "b-" + bot.ID.String(),
				ConversationId: channelID,
				Username:       bot.Name,
				Content:        bot.Profile,
				Avatar:         bot.GetAvatarURL(),
			})
		}
		var msgs []*chatdb.Message
		if afterID != "" {
			msgs, err = s.data.GetChannelMessagesAfter(ctx, channel, afterID)
		} else {
			msgs, err = s.data.GetChannelMessages(ctx, channel)
		}
		if err != nil {
			return errors.Wrap(err, "get channel messages")
		}
		usersByID := fns.ToMap(users, func(user *chatdb.User) uuid.UUID { return user.ID })
		for _, msg := range msgs {
			userID := "Unknown"
			if user, ok := usersByID[msg.AuthorID]; ok {
				userID = user.ProviderID
			}
			client.SendMessage(&chat.ClientMessage{
				ID:             msg.ID.String(),
				Type:           "message",
				UserId:         userID,
				ConversationId: channelID,
				Content:        msg.Content,
			})
		}
		return nil
	}
}

func (s *Service) handleClientMessage(ctx context.Context, clientMsg *chat.ClientMessage) error {
	if clientMsg.Type == "reconnect" && clientMsg.Client != nil {
		for _, channel := range clientMsg.Conversations {
			err := s.sendChannelHistory(ctx, channel.ID, channel.LastMessageID, clientMsg.Client)
			if err != nil {
				return errors.Wrap(err, "send channel history")
			}
		}
		return nil
	} else if clientMsg.Type == "join" && clientMsg.Client != nil {
		return s.sendChannelHistory(ctx, clientMsg.ConversationId, "", clientMsg.Client)
	} else if clientMsg.Type != "message" {
		return nil
	}
	var botID uuid.UUID
	if id, ok := strings.CutPrefix(clientMsg.UserId, "b-"); ok {
		botID, _ = uuid.FromString(id)
	}
	_, err := provider.InboxTopic.Publish(ctx, &provider.Message{
		Provider:   chatdb.ProviderEncorechat,
		ProviderID: clientMsg.ID,
		ChannelID:  clientMsg.ConversationId,
		Author: provider.User{
			ID:    clientMsg.UserId,
			Name:  clientMsg.UserId,
			BotID: botID,
		},
		Content: clientMsg.Content,
		Time:    time.Now(),
		Type:    clientMsg.Type,
	})
	return errors.Wrap(err, "publish message")
}

//encore:api private method=POST path=/encorechat/channels/:channelID/bots/:botID
func (s *Service) SendTyping(ctx context.Context, channelID string, botID uuid.UUID) error {
	s.hub.BroadCast(ctx, &chat.ClientMessage{
		Type:           "typing",
		UserId:         "b-" + botID.String(),
		ConversationId: channelID,
	})
	return nil
}

//encore:api private method=POST path=/encorechat/channels/:channelID/messages
func (s *Service) SendMessage(ctx context.Context, channelID string, req *provider.SendMessageRequest) error {
	if req.Bot == nil {
		return errors.New("only bots can send messages")
	}
	s.hub.BroadCast(ctx, &chat.ClientMessage{
		ID:             req.ID,
		Type:           req.Type,
		UserId:         "b-" + req.Bot.ID.String(),
		ConversationId: channelID,
		Content:        req.Content,
	})
	return nil
}
