package client

import (
	"context"
	"time"

	botdb "encore.app/bot/db"
	"encore.app/chat/service/db"
	"encore.dev/types/uuid"
)

type UserID = string

type ChannelID = string

// Message is a message sent by a user in a provider channel
type Message struct {
	Provider   db.Provider
	ProviderID string
	ChannelID  ChannelID
	Author     User
	Content    string
	Time       time.Time
	Type       string
}

// User is a user in a provider
type User struct {
	ID      UserID
	Name    string
	Profile string
	BotID   uuid.UUID
}

// ChannelInfo is information about a channel in a provider
type ChannelInfo struct {
	Provider db.Provider
	ID       ChannelID
	Name     string
}

// Client is a generic interface implemented by all chat providers
type Client interface {
	// ListChannels lists all channels in the provider
	ListChannels(ctx context.Context) ([]ChannelInfo, error)
	// GetChannelClient gets a client for a specific channel
	GetChannelClient(ctx context.Context, id ChannelID) ChannelClient
	// GetUser gets a user by ID
	GetUser(ctx context.Context, id UserID) (*User, error)
}

// ChannelClient is a client for a specific channel in a provider
type ChannelClient interface {
	// Send sends a message to the channel
	Send(ctx context.Context, bot *botdb.Bot, content string) error
	// ListMessages lists messages in the channel
	ListMessages(ctx context.Context, from *db.Message) ([]*Message, error)
	// GetInfo gets information about the channel
	Info(ctx context.Context) (ChannelInfo, error)
	// Join joins the bot to the channel
	Join(ctx context.Context, bot *botdb.Bot) error
	// Leave leaves the bot from the channel
	Leave(ctx context.Context, bot *botdb.Bot) error
}
