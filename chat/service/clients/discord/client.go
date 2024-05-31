package discord

import (
	"context"

	"github.com/cockroachdb/errors"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider/discord"
	"encore.app/chat/service/clients"
	chatdb "encore.app/chat/service/db"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct{}

func (p *Client) ListChannels(ctx context.Context) ([]client.ChannelInfo, error) {
	resp, err := discord.ListChannels(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}
	return resp.Channels, nil
}

func (p *Client) GetUser(ctx context.Context, id client.UserID) (*client.User, error) {
	return discord.GetUser(ctx, id)
}

func (p *Client) GetChannel(ctx context.Context, id client.ChannelID) client.Channel {
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
	return discord.SendMessage(ctx, c.channelID, &discord.SendMessageRequest{Content: content, Bot: bot})
}

func (c *Channel) ListMessages(ctx context.Context, from *chatdb.Message) ([]*client.Message, error) {
	fromID := ""
	if from != nil {
		fromID = from.ProviderID
	}
	resp, err := discord.ListMessages(ctx, c.channelID, &discord.ListMessagesRequest{FromMessageID: fromID})
	if err != nil {
		return nil, errors.Wrap(err, "list messages")
	}
	return resp.Messages, nil
}

func (c *Channel) Info(ctx context.Context) (client.ChannelInfo, error) {
	return discord.ChannelInfo(ctx, c.channelID)
}

func (c *Channel) Join(ctx context.Context, bot *botdb.Bot) error {
	return discord.JoinChannel(ctx, c.channelID, bot)
}

func (c *Channel) Leave(ctx context.Context, bot *botdb.Bot) error {
	return discord.LeaveChannel(ctx, c.channelID, bot)
}

var (
	_ client.Client  = (*Client)(nil)
	_ client.Channel = (*Channel)(nil)
)
