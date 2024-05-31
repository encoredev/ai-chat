// The LLM service is responsible for processing tasks related to the AI. It receives tasks from the chat service
// and forwards requests to LLM providers.
package llm

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/nfnt/resize"
	"gopkg.in/yaml.v2"

	botdb "encore.app/bot/db"
	chatdb "encore.app/chat/service/db"
	"encore.app/llm/provider"
	"encore.app/llm/service/clients"
	"encore.app/llm/service/clients/gemini"
	"encore.app/llm/service/clients/openai"
	"encore.dev/rlog"
)

//go:embed prompts/reply.txt
var replyPrompt []byte

//go:embed prompts/intro.txt
var introPrompt []byte

//go:embed prompts/goodbye.txt
var goodbyePrompt []byte

//go:embed prompts/avatar.txt
var avatarPrompt []byte

//go:embed prompts/persona.txt
var personaPrompt []byte

//go:embed prompts/response.txt
var responsePrompt []byte

type ChatRequest struct {
	Bots        []*botdb.Bot
	Users       []*chatdb.User
	Channel     *chatdb.Channel
	Messages    []*chatdb.Message
	AdminPrompt string
	Provider    string
}

type BotResponse struct {
	TaskType TaskType
	Channel  *chatdb.Channel
	Messages []*provider.BotMessage
}

//encore:service
type Service struct {
	providers map[string]client.Client
}

func initService() (*Service, error) {
	svc := &Service{
		providers: map[string]client.Client{
			"openai": openai.NewClient(),
			"gemini": gemini.NewClient(),
		},
	}
	return svc, nil
}

func (svc *Service) ProcessTask(ctx context.Context, task *Task) error {
	var res *BotResponse
	var err error
	switch task.Type {
	case TaskTypeJoin:
		res, err = svc.Introduce(ctx, task.Request)
		if err != nil {
			return errors.Wrap(err, "introduce")
		}
	case TaskTypeContinue:
		res, err = svc.ContinueChat(ctx, task.Request)
		if err != nil {
			return errors.Wrap(err, "continue chat")
		}
	case TaskTypeLeave:
		res, err = svc.Goodbye(ctx, task.Request)
		if err != nil {
			return errors.Wrap(err, "goodbye")
		}
	case TaskTypeInstruct:
		res, err = svc.Instruct(ctx, task.Request)
		if err != nil {
			return errors.Wrap(err, "instruct")
		}
	}
	if len(res.Messages) > 0 {
		_, err := BotMessageTopic.Publish(ctx, res)
		if err != nil {
			rlog.Warn("publish message", "error", err)
		}
	}
	return nil
}

//encore:api private path=/ai/instruct
func (svc *Service) Instruct(ctx context.Context, req *provider.ChatRequest) (*BotResponse, error) {
	msgs, err := svc.continueChat(ctx, req)
	if err != nil {
		return nil, err
	}
	return &BotResponse{Messages: msgs, Channel: req.Channel}, nil
}

//encore:api private path=/ai/chat
func (svc *Service) ContinueChat(ctx context.Context, req *provider.ChatRequest) (*BotResponse, error) {
	req.SystemMsg = req.SystemMsg + string(replyPrompt)
	msgs, err := svc.continueChat(ctx, req)
	if err != nil {
		return nil, err
	}
	return &BotResponse{Messages: msgs, Channel: req.Channel}, nil
}

//encore:api private path=/ai/introduce
func (svc *Service) Introduce(ctx context.Context, req *provider.ChatRequest) (*BotResponse, error) {
	req.SystemMsg = req.SystemMsg + fmt.Sprintf(string(introPrompt), req.Channel.Name)
	resp, err := svc.continueChat(ctx, req)
	if err != nil {
		return nil, err
	}
	return &BotResponse{Messages: resp, Channel: req.Channel}, nil
}

//encore:api private path=/ai/goodbye
func (svc *Service) Goodbye(ctx context.Context, req *provider.ChatRequest) (*BotResponse, error) {
	req.SystemMsg = req.SystemMsg + fmt.Sprintf(string(goodbyePrompt), req.Channel)
	resp, err := svc.continueChat(ctx, req)
	if err != nil {
		return nil, err
	}
	return &BotResponse{Messages: resp, Channel: req.Channel}, nil
}

//go:embed prompts/create_persona.txt
var botPrompt []byte

type GenerateBotProfileRequest struct {
	Name     string `json:"name"`
	Prompt   string `json:"prompt"`
	Provider string `json:"provider"`
}

type GenerateBotResponse struct {
	Profile string `json:"profile"`
	Avatar  []byte
}

//encore:api private method=POST path=/ai/bot
func (svc *Service) GenerateBotProfile(ctx context.Context, req *GenerateBotProfileRequest) (*GenerateBotResponse, error) {
	prov, ok := svc.providers[req.Provider]
	if !ok {
		return nil, errors.Newf("provider not found: %s", req.Provider)
	}
	resp, err := prov.Ask(ctx, fmt.Sprintf(string(botPrompt), req.Name, req.Prompt))
	if err != nil {
		return nil, errors.Wrap(err, "ask")
	}
	img, err := svc.generateAvatar(ctx, req.Provider, resp)
	if err != nil {
		return nil, errors.Wrap(err, "generate avatar")
	}
	return &GenerateBotResponse{Profile: resp, Avatar: img}, nil
}

func formatBotProfiles(bots []*botdb.Bot) string {
	res := strings.Builder{}
	for _, b := range bots {
		res.WriteString(b.Name)
		res.WriteString(": ")
		res.WriteString(b.Profile)
		res.WriteString("\n")
	}
	return res.String()
}

func formatResponsePrompt(bots []*botdb.Bot) string {
	users := strings.Builder{}
	for i, user := range bots {
		if i > 0 {
			users.WriteString(", ")
		}
		users.WriteString(user.Name)
	}
	names := strings.TrimSuffix(users.String(), ", ")
	return fmt.Sprintf(string(responsePrompt), names)
}

func (svc *Service) continueChat(ctx context.Context, req *provider.ChatRequest) ([]*provider.BotMessage, error) {
	prov, ok := svc.providers[req.Provider]
	if !ok {
		return nil, errors.Newf("provider not found: %s", req.Provider)
	}
	botByName := make(map[string]*botdb.Bot)
	for _, b := range req.Bots {
		botByName[b.Name] = b
	}
	var messages []*provider.BotMessage
	req.Messages = append(req.Messages, &chatdb.Message{
		ChannelID: req.Channel.ID,
		AuthorID:  chatdb.Admin.ID,
		Content:   req.SystemMsg + formatResponsePrompt(req.Bots),
		Timestamp: time.Now().UTC(),
	})
	req.SystemMsg = fmt.Sprintf(string(personaPrompt), formatBotProfiles(req.Bots))
	resp, err := prov.ContinueChat(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "continue chat")
	}
	rlog.Info("AI response", "response", resp)
	_, after, ok := strings.Cut(resp, "```yaml")
	if ok {
		resp = after
		before, _, ok := strings.Cut(resp, "```")
		if ok {
			resp = before
		}
	}
	respMap := make(map[string]string)
	err = yaml.Unmarshal([]byte(resp), &respMap)
	if err != nil {
		return nil, err
	}

	for botName, content := range respMap {
		botName := strings.Split(botName, "/")
		if botName[len(botName)-1] == "None" {
			continue
		}
		bot, ok := botByName[botName[len(botName)-1]]
		if !ok {
			rlog.Warn("bot not found", "bot", botName)
			continue
		}
		messages = append(messages, &provider.BotMessage{
			Bot:     bot,
			Content: content,
			Time:    time.Now(),
		})
	}
	return messages, nil
}

func (svc *Service) generateAvatar(ctx context.Context, provider, prompt string) ([]byte, error) {
	prov, ok := svc.providers[provider]
	if !ok {
		return nil, errors.Wrap(errors.New("provider not found"), "generate avatar")
	}
	img, err := prov.GenerateAvatar(ctx, fmt.Sprintf(string(avatarPrompt), prompt))
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, nil
	}
	if img.Bounds().Dx() > 256 {
		img = resize.Resize(256, 0, img, resize.Lanczos3)
	}
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, img)
	if err != nil {
		return nil, errors.Wrap(err, "encode image")
	}
	return buffer.Bytes(), nil
}
