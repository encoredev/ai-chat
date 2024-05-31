// Discord service provides functionality for interacting with Discord channels and users.
// It implements the chat provider API
package discord

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"os/signal"
	"strings"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/cockroachdb/errors"
	"github.com/nfnt/resize"

	botdb "encore.app/bot/db"
	"encore.app/chat/provider"
	"encore.app/chat/provider/discord/db"
	"encore.app/chat/service/clients"
	chatdb "encore.app/chat/service/db"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
)

var discorddb = sqldb.NewDatabase("discord", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})

var secrets struct {
	DiscordToken string
}

//encore:service
type Service struct {
	client *discord.Session
}

func (s *Service) handleMessage(ctx context.Context, msg *discord.MessageCreate) error {
	message := ToProviderMessage(msg.Message)
	if message == nil {
		return nil
	}
	_, err := provider.MessageTopic.Publish(ctx, message)
	return errors.Wrap(err, "publish message")
}

func initService() (*Service, error) {
	client, err := discord.New("Bot " + secrets.DiscordToken)
	if err != nil {
		return nil, err
	}
	svc := &Service{client: client}
	err = svc.subscribeToMessages(context.Background(), svc.handleMessage)
	if err != nil {
		return nil, errors.Wrap(err, "subscribe to messages")
	}
	return svc, nil
}

type ListChannelsResponse struct {
	Channels []client.ChannelInfo
}

type DiscordAuthRequest struct {
	Code        string `json:"code"`
	GuildID     string `json:"guild_id"`
	Permissions int64  `json:"permissions"`
}

//encore:api public method=GET path=/discord/oauth
func (p *Service) AuthURL(ctx context.Context, req *DiscordAuthRequest) error {
	// Encrypt and store the token somewhere
	return nil
}

//encore:api private method=GET path=/discord/channels
func (p *Service) ListChannels(ctx context.Context) (*ListChannelsResponse, error) {
	guilds, err := p.client.UserGuilds(100, "", "", false)
	if err != nil {
		return nil, errors.Wrap(err, "error getting guilds")
	}
	var channelInfos []client.ChannelInfo
	for _, guild := range guilds {
		channels, err := p.client.GuildChannels(guild.ID)
		if err != nil {
			return nil, errors.Wrap(err, "error getting guilds")
		}
		for _, channel := range channels {
			channelInfos = append(channelInfos, client.ChannelInfo{
				Provider: chatdb.ProviderDiscord,
				ID:       channel.ID,
				Name:     channel.Name,
			})
		}
	}
	return &ListChannelsResponse{Channels: channelInfos}, nil
}

func (p *Service) subscribeToMessages(ctx context.Context, fn func(ctx context.Context, msg *discord.MessageCreate) error) error {
	p.client.AddHandler(func(sess *discord.Session, msg *discord.MessageCreate) {
		err := fn(ctx, msg)
		if err != nil {
			rlog.Error("error handling message: %v", err)
		}
	})

	p.client.Identify.Intents = discord.IntentGuildMessages | discord.IntentMessageContent

	// Open the websocket and begin listening.
	err := p.client.Open()
	if err != nil {
		return errors.Wrap(err, "error opening Discord session")
	}
	rlog.Info("Discord subscription is running.")

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		_ = p.client.Close()
	}()
	return nil
}

//encore:api private method=GET path=/discord/users/:userID
func (p *Service) GetUser(ctx context.Context, userID string) (*client.User, error) {
	_, username, isBot := strings.Cut(userID, ":")
	if isBot {
		return &client.User{
			ID:   userID,
			Name: username,
		}, nil
	}
	user, err := p.client.User(userID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}
	return &client.User{
		ID:   userID,
		Name: user.GlobalName,
	}, nil
}

//encore:api private method=POST path=/discord/channels/:channelID/leave
func (c *Service) LeaveChannel(ctx context.Context, channelID string, bot *botdb.Bot) error {
	q := db.New()
	webhook, err := q.GetWebhookForBot(ctx, discorddb.Stdlib(), db.GetWebhookForBotParams{
		Channel: channelID,
		BotID:   bot.ID,
	})
	if err != nil {
		return errors.Wrap(err, "error getting webhook")
	}
	err = c.client.WebhookDelete(webhook.ProviderID)
	if err != nil {
		return errors.Wrap(err, "error deleting webhook")
	}
	_, err = q.DeleteWebhook(ctx, discorddb.Stdlib(), webhook.ID)
	if err != nil {
		return errors.Wrap(err, "error deleting webhook")
	}
	return nil
}

func generateAvatarDataURI(data []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, "error decoding image")
	}
	if img.Bounds().Dx() > 256 || img.Bounds().Dy() > 256 {
		img = resize.Resize(128, 128, img, resize.Lanczos3)
		buffer := new(bytes.Buffer)
		err = png.Encode(buffer, img)
		if err != nil {
			return "", errors.Wrap(err, "error encoding image")
		}
		data = buffer.Bytes()
	}
	b64Data := base64.StdEncoding.EncodeToString(data)
	return "data:image/png;base64," + b64Data, nil
}

//encore:api private method=POST path=/discord/channels/:channelID/join
func (c *Service) JoinChannel(ctx context.Context, channelID string, bot *botdb.Bot) error {
	var err error
	avatarURI := ""
	if len(bot.Avatar) > 0 {
		avatarURI, err = generateAvatarDataURI(bot.Avatar)
		if err != nil {
			return errors.Wrap(err, "error generating avatar data URI")
		}
	}
	q := db.New()
	hook, err := q.GetWebhookForBot(ctx, discorddb.Stdlib(), db.GetWebhookForBotParams{
		Channel: channelID,
		BotID:   bot.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		hook, err := c.client.WebhookCreate(channelID, bot.Name, avatarURI)
		if err != nil {
			return errors.Wrap(err, "error creating webhook")
		}
		_, err = q.InsertWebhook(ctx, discorddb.Stdlib(), db.InsertWebhookParams{
			ProviderID: hook.ID,
			Channel:    hook.ChannelID,
			Name:       hook.Name,
			Token:      hook.Token,
			BotID:      bot.ID,
		})
		if err != nil {
			return errors.Wrap(err, "error inserting webhook")
		}
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting webhook")
	}
	_, err = c.client.WebhookEdit(hook.ProviderID, bot.Name, avatarURI, channelID)
	if err != nil {
		return errors.Wrap(err, "error editing webhook")
	}
	return nil
}

type SendMessageRequest struct {
	Content string
	Bot     *botdb.Bot
}

//encore:api private method=POST path=/discord/channels/:channelID/messages
func (c *Service) SendMessage(ctx context.Context, channelID string, req *SendMessageRequest) error {
	webhook, err := db.New().GetWebhookForBot(ctx, discorddb.Stdlib(), db.GetWebhookForBotParams{
		Channel: channelID,
		BotID:   req.Bot.ID,
	})
	if err != nil {
		return errors.Wrap(err, "error getting webhook")
	}
	_, err = c.client.WebhookExecute(webhook.ProviderID, webhook.Token, false, &discord.WebhookParams{
		Content:  req.Content,
		Username: req.Bot.Name,
	})
	return errors.Wrap(err, "error sending message")
}

func ToProviderMessage(msg *discord.Message) *client.Message {
	if msg.Content == "" || msg.Type != discord.MessageTypeDefault {
		return nil
	}
	author := client.User{
		ID:   msg.Author.ID,
		Name: msg.Author.Username,
	}
	if msg.Author.Bot {
		author.ID = msg.Author.ID + ":" + msg.Author.Username
		hook, err := db.New().GetWebhookByID(context.Background(), discorddb.Stdlib(), msg.Author.ID)
		if err == nil {
			author.BotID = hook.BotID
		}
	}
	return &client.Message{
		Provider:   chatdb.ProviderDiscord,
		ProviderID: msg.ID,
		Channel:    msg.ChannelID,
		Author:     author,
		Content:    msg.Content,
		Time:       msg.Timestamp.UTC(),
	}
}

type ListMessagesRequest struct {
	FromMessageID string
}

type ListMessagesResponse struct {
	Messages []*client.Message
}

//encore:api private method=GET path=/discord/channels/:channelID/messages
func (c *Service) ListMessages(ctx context.Context, channelID string, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	msgs, err := c.client.ChannelMessages(channelID, 100, "", req.FromMessageID, "")
	if err != nil {
		return nil, errors.Wrap(err, "error getting messages")
	}
	var messages []*client.Message
	for i := len(msgs) - 1; i >= 0; i-- {
		if msg := ToProviderMessage(msgs[i]); msg != nil {
			messages = append(messages, msg)
		}
	}
	return &ListMessagesResponse{Messages: messages}, nil
}

//encore:api private method=GET path=/discord/channels/:channelID/info
func (c *Service) ChannelInfo(ctx context.Context, channelID string) (client.ChannelInfo, error) {
	resp, err := c.client.Channel(channelID)
	if err != nil {
		return client.ChannelInfo{}, errors.Wrap(err, "error getting channel info")
	}
	return client.ChannelInfo{
		Provider: chatdb.ProviderDiscord,
		ID:       client.ChannelID(resp.ID),
		Name:     resp.Name,
	}, nil
}
