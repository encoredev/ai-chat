// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"context"

	"encore.dev/types/uuid"
)

type Querier interface {
	DeleteBot(ctx context.Context, db DBTX, id uuid.UUID) (*Bot, error)
	GetAvatar(ctx context.Context, db DBTX, botID uuid.UUID) (*Avatar, error)
	GetBot(ctx context.Context, db DBTX, id uuid.UUID) (*Bot, error)
	GetBotByName(ctx context.Context, db DBTX, name string) (*Bot, error)
	GetBots(ctx context.Context, db DBTX, ids []uuid.UUID) ([]*Bot, error)
	InsertAvatar(ctx context.Context, db DBTX, arg InsertAvatarParams) error
	InsertBot(ctx context.Context, db DBTX, arg InsertBotParams) (*Bot, error)
	ListBot(ctx context.Context, db DBTX) ([]*Bot, error)
}

var _ Querier = (*Queries)(nil)
