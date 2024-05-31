package encorechat

import (
	"context"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider"
	"encore.app/chat/provider/encorechat"
	"encore.app/chat/service/client"
	chatdb "encore.app/chat/service/db"
)

func NewClient(ctx context.Context) (*Client, bool) {
	if err := encorechat.Ping(ctx); err != nil {
		return nil, false
	}
	return &Client{}, true
}

// Client wraps the discord service endpoints to implement the chat client interface.
type Client struct{}

func (p *Client) ListChannels(ctx context.Context) ([]client.ChannelInfo, error) {
	return nil, nil
}

func (p *Client) GetUser(ctx context.Context, id client.UserID) (*client.User, error) {
	return &client.User{
		ID:   id,
		Name: id,
	}, nil
}

func (p *Client) GetChannelClient(ctx context.Context, id client.ChannelID) client.ChannelClient {
	return &Channel{
		Client:    p,
		channelID: id,
	}
}

type Channel struct {
	*Client
	channelID string
}

func (c *Channel) Send(ctx context.Context, bot *botdb.Bot, content string) error {
	return encorechat.SendMessage(ctx, c.channelID, &provider.SendMessageRequest{Content: content, Bot: bot})
}

func (c *Channel) ListMessages(ctx context.Context, from *chatdb.Message) ([]*client.Message, error) {
	return nil, nil
}

func (c *Channel) Info(ctx context.Context) (client.ChannelInfo, error) {
	return client.ChannelInfo{
		Provider: chatdb.ProviderEncorechat,
		ID:       c.channelID,
		Name:     c.channelID,
	}, nil
}

func (c *Channel) Join(ctx context.Context, bot *botdb.Bot) error {
	return nil
}

func (c *Channel) Leave(ctx context.Context, bot *botdb.Bot) error {
	return nil
}

var (
	_ client.Client        = (*Client)(nil)
	_ client.ChannelClient = (*Channel)(nil)
)
