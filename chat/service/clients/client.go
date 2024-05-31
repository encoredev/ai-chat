package client

import (
	"context"
	"time"

	db2 "encore.app/bot/db"
	"encore.app/chat/service/db"
	"encore.dev/types/uuid"
)

type UserID = string

type ChannelID = string

type Message struct {
	Provider   db.Provider
	ProviderID string
	Channel    ChannelID
	Author     User
	Content    string
	Time       time.Time
}

type User struct {
	ID      UserID
	Name    string
	Profile string
	BotID   uuid.UUID
}

type ChannelInfo struct {
	Provider db.Provider
	ID       ChannelID
	Name     string
}

type Channel interface {
	Send(ctx context.Context, bot *db2.Bot, content string) error
	ListMessages(ctx context.Context, from *db.Message) ([]*Message, error)
	Info(ctx context.Context) (ChannelInfo, error)
	Join(ctx context.Context, bot *db2.Bot) error
	Leave(ctx context.Context, bot *db2.Bot) error
}

type Client interface {
	ListChannels(ctx context.Context) ([]ChannelInfo, error)
	GetChannel(ctx context.Context, id ChannelID) Channel
	GetUser(ctx context.Context, id UserID) (*User, error)
}
