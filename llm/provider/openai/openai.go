// OpenAI is an llm provider implementation for the OpenAI API.
package openai

import (
	"context"
	"encoding/base64"

	"github.com/cockroachdb/errors"
	"github.com/sashabaranov/go-openai"

	"encore.app/llm/provider"
	"encore.dev/config"
)

var secrets struct {
	OpenAIKey string
}

type Config struct {
	ChatModel   config.String
	ImageModel  config.String
	MaxTokens   config.Int
	Temperature config.Float32
	TopP        config.Float32
}

var cfg = config.Load[*Config]()

//encore:service
type Service struct {
	client *openai.Client
}

func initService() (*Service, error) {
	svc := &Service{
		client: openai.NewClient(secrets.OpenAIKey),
	}
	return svc, nil
}

type GenerateAvatarRequest struct {
	Prompt string
}

type GenerateAvatarResponse struct {
	Image []byte
}

//encore:api private method=POST path=/openai/generate-avatar
func (p *Service) GenerateAvatar(ctx context.Context, req *GenerateAvatarRequest) (*GenerateAvatarResponse, error) {
	resp, err := p.client.CreateImage(ctx, openai.ImageRequest{
		Prompt:         req.Prompt,
		Model:          cfg.ImageModel(),
		N:              1,
		Quality:        "standard",
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create image")
	}
	data, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, errors.Wrap(err, "decode image")
	}
	return &GenerateAvatarResponse{Image: data}, nil
}

type AskRequest struct {
	Message string
}

type AskResponse struct {
	Message string
}

//encore:api private method=POST path=/openai/ask
func (p *Service) Ask(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: cfg.ChatModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Message,
			},
		},
		MaxTokens:   cfg.MaxTokens(),
		N:           1,
		Temperature: cfg.Temperature(),
		TopP:        cfg.TopP(),
	})
	if err != nil {
		return nil, err
	}
	return &AskResponse{Message: resp.Choices[0].Message.Content}, nil
}

type ContinueChatResponse struct {
	Message string
}

//encore:api private method=POST path=/openai/continue-chat
func (p *Service) ContinueChat(ctx context.Context, req *provider.ChatRequest) (*ContinueChatResponse, error) {
	var messages []openai.ChatCompletionMessage
	if req.SystemMsg != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemMsg,
		})
	}
	for _, m := range req.Messages {
		role := openai.ChatMessageRoleUser
		if req.FromBot(m) {
			// The bot is always the assistant in this case and should reply without prefixes
			role = openai.ChatMessageRoleAssistant
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: req.Format(m),
		})
	}
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     cfg.ChatModel(),
		Messages:  messages,
		MaxTokens: cfg.MaxTokens(),
		N:         1,
	})
	if err != nil {
		return nil, err
	}
	return &ContinueChatResponse{Message: resp.Choices[0].Message.Content}, nil
}
