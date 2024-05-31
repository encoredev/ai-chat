package slack

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider/slack"
	"encore.app/chat/service/client"
	chatdb "encore.app/chat/service/db"
)

func NewClient(ctx context.Context) (*Client, bool) {
	if slack.Ping(ctx) != nil {
		return nil, false
	}
	return &Client{}, true
}

// Client wraps the slack service endpoints to implement the chat client interface.
type Client struct{}

func (s *Client) ListChannels(ctx context.Context) ([]client.ChannelInfo, error) {
	resp, err := slack.ListChannels(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}
	return resp.Channels, nil
}

func (s *Client) GetUser(ctx context.Context, id client.UserID) (*client.User, error) {
	return slack.GetUser(ctx, id)
}

func (s *Client) GetChannelClient(ctx context.Context, id client.ChannelID) client.ChannelClient {
	return &Channel{
		channelID: id,
	}
}

type Channel struct {
	channelID client.ChannelID
}

func (c *Channel) Send(ctx context.Context, bot *botdb.Bot, content string) error {
	return slack.SendMessage(ctx, c.channelID, &slack.SendMessageRequest{Content: content, Bot: bot})
}

func (c *Channel) ListMessages(ctx context.Context, from *chatdb.Message) ([]*client.Message, error) {
	fromTimestamp := ""
	if from != nil {
		fromTimestamp = fmt.Sprintf("%f", float64(from.Timestamp.UnixMicro()/1e6))
	}
	resp, err := slack.ListMessages(ctx, c.channelID, &slack.ListMessagesRequest{FromTimestamp: fromTimestamp})
	if err != nil {
		return nil, errors.Wrap(err, "list messages")
	}
	return resp.Messages, nil
}

func (c *Channel) Info(ctx context.Context) (client.ChannelInfo, error) {
	return slack.ChannelInfo(ctx, c.channelID)
}

func (c *Channel) Join(ctx context.Context, bot *botdb.Bot) error {
	return slack.JoinChannel(ctx, c.channelID, bot)
}

func (c *Channel) Leave(ctx context.Context, bot *botdb.Bot) error {
	return slack.LeaveChannel(ctx, c.channelID, bot)
}

var (
	_ client.Client        = (*Client)(nil)
	_ client.ChannelClient = (*Channel)(nil)
)
