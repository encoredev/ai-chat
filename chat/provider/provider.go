package provider

import (
	"encore.app/bot/db"
	"encore.app/chat/service/client"
)

type ListChannelsResponse struct {
	Channels []client.ChannelInfo
}

type SendMessageRequest struct {
	Content string
	Bot     *db.Bot
	UserID  string
	Type    string
}

type ListMessagesResponse struct {
	Messages []*client.Message
}

type ListMessagesRequest struct {
	FromTimestamp string
	FromMessageID string
}
