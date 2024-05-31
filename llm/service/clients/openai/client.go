package openai

import (
	"bytes"
	"context"
	"image"

	"github.com/cockroachdb/errors"

	"encore.app/llm/provider"
	"encore.app/llm/provider/openai"
	"encore.app/llm/service/clients"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct{}

func (p Client) ContinueChat(ctx context.Context, req *provider.ChatRequest) (string, error) {
	resp, err := openai.ContinueChat(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "continue chat")
	}
	return resp.Message, nil
}

func (p Client) Ask(ctx context.Context, msg string) (string, error) {
	resp, err := openai.Ask(ctx, &openai.AskRequest{
		Message: msg,
	})
	if err != nil {
		return "", errors.Wrap(err, "ask")
	}
	return resp.Message, nil
}

func (p Client) GenerateAvatar(ctx context.Context, prompt string) (image.Image, error) {
	resp, err := openai.GenerateAvatar(ctx, &openai.GenerateAvatarRequest{
		Prompt: prompt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "generate avatar")
	}
	img, _, err := image.Decode(bytes.NewReader(resp.Image))
	if err != nil {
		return nil, errors.Wrap(err, "decode image")
	}
	return img, nil
}

var _ client.Client = (*Client)(nil)
