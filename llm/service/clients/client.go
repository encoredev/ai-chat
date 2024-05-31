package client

import (
	"context"
	"image"

	"encore.app/llm/provider"
)

type Client interface {
	ContinueChat(ctx context.Context, req *provider.ChatRequest) (string, error)
	Ask(ctx context.Context, msg string) (string, error)
	GenerateAvatar(ctx context.Context, prompt string) (image.Image, error)
}
