package gemini

import (
	"context"
	"image"

	"github.com/cockroachdb/errors"

	"encore.app/llm/provider"
	"encore.app/llm/provider/gemini"
	"encore.app/llm/service/clients"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct{}

func (p *Client) ContinueChat(ctx context.Context, req *provider.ChatRequest) (string, error) {
	resp, err := gemini.ContinueChat(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "continue chat")
	}
	return resp.Message, nil
}

func (p *Client) Ask(ctx context.Context, msg string) (string, error) {
	resp, err := gemini.Ask(ctx, &gemini.AskRequest{
		Message: msg,
	})
	if err != nil {
		return "", errors.Wrap(err, "ask")
	}
	return resp.Message, nil
}

func (p *Client) GenerateAvatar(ctx context.Context, prompt string) (image.Image, error) {
	// Imagen is not yet available
	return nil, nil
}

var _ client.Client = (*Client)(nil)
