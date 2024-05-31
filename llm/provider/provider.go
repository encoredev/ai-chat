package provider

import (
	_ "embed"
	"fmt"
	"time"

	botdb "encore.app/bot/db"
	chatdb "encore.app/chat/service/db"
	"encore.dev/types/uuid"
)

type BotMessage struct {
	Bot     *botdb.Bot
	Content string
	Time    time.Time
}

type ChatRequest struct {
	Bots      []*botdb.Bot
	Users     []*chatdb.User
	Messages  []*chatdb.Message
	Channel   *chatdb.Channel
	SystemMsg string
	Provider  string

	botsByID  map[uuid.UUID]*botdb.Bot
	usersByID map[uuid.UUID]*chatdb.User
}

var unknownUser = &chatdb.User{
	Name: "Unknown User",
}

func (req *ChatRequest) Format(msg *chatdb.Message) string {
	// Admin messages should not be formated
	if msg.AuthorID == uuid.Nil {
		return fmt.Sprintf("Admin: %s", msg.Content)
	}
	user, bot := req.UserForMessage(msg)
	name := user.Name
	if bot != nil {
		name = bot.Name
	}
	return fmt.Sprintf("%s %s/%s: %s", msg.Timestamp.Format("01-02 15:04"), req.Channel.Name, name, msg.Content)
}

func (req *ChatRequest) FromBot(msg *chatdb.Message) bool {
	_, bot := req.UserForMessage(msg)
	return bot != nil
}

func (req *ChatRequest) UserForMessage(msg *chatdb.Message) (*chatdb.User, *botdb.Bot) {
	user, ok := req.UsersByID()[msg.AuthorID]
	if !ok {
		return unknownUser, nil
	}
	if user.BotID == nil {
		return user, nil
	}
	bot, ok := req.BotsByID()[*user.BotID]
	if !ok {
		return user, nil
	}
	return user, bot
}

func (req *ChatRequest) BotsByID() map[uuid.UUID]*botdb.Bot {
	if req.botsByID != nil {
		return req.botsByID
	}
	req.botsByID = make(map[uuid.UUID]*botdb.Bot)
	for _, b := range req.Bots {
		req.botsByID[b.ID] = b
	}
	return req.botsByID
}

func (req *ChatRequest) UsersByID() map[uuid.UUID]*chatdb.User {
	if req.usersByID != nil {
		return req.usersByID
	}
	req.usersByID = make(map[uuid.UUID]*chatdb.User)
	for _, u := range req.Users {
		req.usersByID[u.ID] = u
	}
	return req.usersByID
}
