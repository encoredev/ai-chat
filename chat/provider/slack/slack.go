// Slack service provides functionality for interacting with slack channels and users.
// It implements the chat provider API
package slack

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/slack-go/slack"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider"
	"encore.app/chat/provider/slack/db"
	"encore.app/chat/service/clients"
	chatdb "encore.app/chat/service/db"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

const (
	BotIDPayload        = "bot_id"
	BotMessageEventType = "bot_message"
)

var slackdb = sqldb.NewDatabase("slack", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})

// This uses Encore's built-in secrets manager, learn more: https://encore.dev/docs/primitives/secrets
var secrets struct {
	SlackToken string
}

//encore:service
type Service struct {
	client *slack.Client
	botID  string
}

func initService() (*Service, error) {
	client := slack.New(secrets.SlackToken)
	resp, err := client.AuthTest()
	if err != nil {
		return nil, errors.Wrap(err, "auth test")
	}
	return &Service{
		client: client,
		botID:  resp.BotID,
	}, nil
}

type SlackEvent struct {
	Token     string          `json:"token"`
	Challenge string          `json:"challenge"`
	Type      string          `json:"type"`
	Event     json.RawMessage `json:"event"`
}

type ChallengeResponse struct {
	Challenge string `json:"challenge"`
}

//encore:api private path=/slack/message
func (svc *Service) WebhookMessage(ctx context.Context, req *SlackEvent) (*ChallengeResponse, error) {
	switch req.Type {
	case "url_verification":
		return &ChallengeResponse{
			Challenge: req.Challenge,
		}, nil
	case "event_callback":
		msg, err := svc.ParseMessage(req.Event)
		if err != nil {
			return nil, err
		}
		if msg == nil {
			return nil, nil
		}

		_, err = provider.MessageTopic.Publish(ctx, msg)
		if err != nil {
			return nil, errors.Wrap(err, "publish message")
		}
	}
	return nil, nil
}

type ListChannelsResponse struct {
	Channels []client.ChannelInfo
}

//encore:api private method=GET path=/slack/channels
func (s *Service) ListChannels(ctx context.Context) (*ListChannelsResponse, error) {
	resp, _, err := s.client.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "mpim", "im"},
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}
	var rtn []client.ChannelInfo
	for _, channel := range resp {
		rtn = append(rtn, client.ChannelInfo{
			Provider: chatdb.ProviderSlack,
			ID:       channel.ID,
			Name:     channel.Name,
		})
	}
	return &ListChannelsResponse{Channels: rtn}, nil
}

//encore:api private method=GET path=/slack/users/:userID
func (s *Service) GetUser(ctx context.Context, userID string) (*client.User, error) {
	if strings.HasPrefix(userID, "B") {
		return nil, nil
	}
	user, err := s.client.GetUserInfo(userID)
	if err != nil {
		return nil, errors.Wrapf(err, "get user info: %s", userID)
	}
	name := user.Name
	if user.Profile.DisplayName != "" {
		name = user.Profile.DisplayName
	}
	return &client.User{
		ID:      userID,
		Name:    name,
		Profile: user.Profile.Title,
	}, nil
}

type Message struct {
	slack.Msg
	Blocks json.RawMessage `json:"blocks"`
}

func (s *Service) ParseMessage(data json.RawMessage) (*client.Message, error) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return s.toProviderMessage(msg.Msg, msg.Channel), nil

}

//encore:api private method=POST path=/slack/channels/:channelID/leave
func (s *Service) LeaveChannel(ctx context.Context, channelID string, bot *botdb.Bot) error {
	_, err := s.client.LeaveConversationContext(ctx, channelID)
	if err != nil {
		return errors.Wrap(err, "leave conversation")
	}
	_, err = db.New().DeleteAvatar(ctx, slackdb.Stdlib(), bot.Name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "leave conversation")
	}
	return nil
}

//encore:api private method=POST path=/slack/channels/:channelID/join
func (s *Service) JoinChannel(ctx context.Context, channelID string, bot *botdb.Bot) error {
	_, _, _, err := s.client.JoinConversationContext(ctx, channelID)
	return errors.Wrap(err, "join conversation")
}

//encore:api private method=GET path=/slack/channels/:channelID
func (s *Service) ChannelInfo(ctx context.Context, channelID string) (client.ChannelInfo, error) {
	resp, err := s.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return client.ChannelInfo{}, err
	}
	return client.ChannelInfo{
		Provider: chatdb.ProviderSlack,
		ID:       channelID,
		Name:     resp.Name,
	}, nil
}

type SendMessageRequest struct {
	Content string
	Bot     *botdb.Bot
}

//encore:api private method=POST path=/slack/channels/:channelID/messages
func (s *Service) SendMessage(ctx context.Context, channelID string, req *SendMessageRequest) error {
	avatar := req.Bot.GetAvatarURL()
	_, _, err := s.client.PostMessageContext(
		ctx,
		channelID,
		slack.MsgOptionMetadata(slack.SlackMetadata{
			EventType: BotMessageEventType,
			EventPayload: map[string]interface{}{
				BotIDPayload: req.Bot.ID,
			},
		}),
		slack.MsgOptionUsername(req.Bot.Name),
		slack.MsgOptionText(req.Content, false),
		slack.MsgOptionIconURL(avatar))
	return errors.Wrap(err, "post message")
}

var acceptedSubTypes = []string{
	"",
	"bot_message",
}

type ListMessagesResponse struct {
	Messages []*client.Message
}

type ListMessagesRequest struct {
	FromTimestamp string
}

//encore:api private method=GET path=/slack/channels/:channelID/messages
func (s *Service) ListMessages(ctx context.Context, channelID string, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	resp, err := s.client.GetConversationHistoryContext(ctx, &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Oldest:    req.FromTimestamp,
		Limit:     100,
	})
	if err != nil {
		return nil, err
	}
	var rtn []*client.Message
	for i := len(resp.Messages) - 1; i >= 0; i-- {
		msg := s.toProviderMessage(resp.Messages[i].Msg, channelID)
		if msg == nil {
			continue
		}
		rtn = append(rtn, msg)
	}
	return &ListMessagesResponse{Messages: rtn}, nil
}

func (svc *Service) toProviderMessage(msg slack.Msg, channel client.ChannelID) *client.Message {
	if msg.Text == "" || msg.Type != "message" || msg.Hidden || !slices.Contains(acceptedSubTypes, msg.SubType) {
		return nil
	}
	author := client.User{
		ID:   msg.User,
		Name: msg.Username,
	}
	if msg.SubType == "bot_message" {
		var err error
		author.ID = msg.BotID + ":" + msg.Username
		if msg.Metadata.EventType == BotMessageEventType {
			botID, hasID := msg.Metadata.EventPayload[BotIDPayload]
			if botIDStr, ok := botID.(string); hasID && ok {
				author.ID = fmt.Sprintf("B-%s", botIDStr)
				author.BotID, err = uuid.FromString(botIDStr)
				if err != nil {
					rlog.Warn("invalid bot id", "id", botIDStr)
				}
			}
		}
	}
	ts, _ := strconv.ParseFloat(msg.Timestamp, 64)
	return &client.Message{
		Provider:   chatdb.ProviderSlack,
		ProviderID: msg.ClientMsgID,
		Channel:    channel,
		Author:     author,
		Content:    msg.Text,
		Time:       time.UnixMicro(int64(ts * 1e6)).UTC(),
	}
}
