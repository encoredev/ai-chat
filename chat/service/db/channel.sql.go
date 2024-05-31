// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: channel.sql

package db

import (
	"context"

	"encore.dev/types/uuid"
)

const getBotChannel = `-- name: GetBotChannel :one
SELECT bot FROM bot_channel WHERE bot = $1 AND channel = $2 AND deleted IS NULL
`

type GetBotChannelParams struct {
	Bot     uuid.UUID
	Channel uuid.UUID
}

func (q *Queries) GetBotChannel(ctx context.Context, db DBTX, arg GetBotChannelParams) (uuid.UUID, error) {
	row := db.QueryRowContext(ctx, getBotChannel, arg.Bot, arg.Channel)
	var bot uuid.UUID
	err := row.Scan(&bot)
	return bot, err
}

const getChannel = `-- name: GetChannel :one
SELECT id, provider_id, provider, name, deleted FROM channel WHERE id = $1 AND deleted IS NULL
`

func (q *Queries) GetChannel(ctx context.Context, db DBTX, id uuid.UUID) (*Channel, error) {
	row := db.QueryRowContext(ctx, getChannel, id)
	var i Channel
	err := row.Scan(
		&i.ID,
		&i.ProviderID,
		&i.Provider,
		&i.Name,
		&i.Deleted,
	)
	return &i, err
}

const getChannelByProviderId = `-- name: GetChannelByProviderId :one
SELECT id, provider_id, provider, name, deleted FROM channel WHERE provider_id = $1 AND provider = $2 AND deleted IS NULL
`

type GetChannelByProviderIdParams struct {
	ProviderID string
	Provider   Provider
}

func (q *Queries) GetChannelByProviderId(ctx context.Context, db DBTX, arg GetChannelByProviderIdParams) (*Channel, error) {
	row := db.QueryRowContext(ctx, getChannelByProviderId, arg.ProviderID, arg.Provider)
	var i Channel
	err := row.Scan(
		&i.ID,
		&i.ProviderID,
		&i.Provider,
		&i.Name,
		&i.Deleted,
	)
	return &i, err
}

const listBotsInChannel = `-- name: ListBotsInChannel :many
SELECT bot FROM bot_channel WHERE channel = $1 AND deleted IS NULL
`

func (q *Queries) ListBotsInChannel(ctx context.Context, db DBTX, channel uuid.UUID) ([]uuid.UUID, error) {
	rows, err := db.QueryContext(ctx, listBotsInChannel, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []uuid.UUID{}
	for rows.Next() {
		var bot uuid.UUID
		if err := rows.Scan(&bot); err != nil {
			return nil, err
		}
		items = append(items, bot)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannels = `-- name: ListChannels :many
SELECT id, provider_id, provider, name, deleted FROM channel WHERE deleted IS NULL
`

func (q *Queries) ListChannels(ctx context.Context, db DBTX) ([]*Channel, error) {
	rows, err := db.QueryContext(ctx, listChannels)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Channel{}
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ID,
			&i.ProviderID,
			&i.Provider,
			&i.Name,
			&i.Deleted,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelsByProvider = `-- name: ListChannelsByProvider :many
SELECT id, provider_id, provider, name, deleted FROM channel WHERE deleted IS NULL AND provider = $1
`

func (q *Queries) ListChannelsByProvider(ctx context.Context, db DBTX, provider Provider) ([]*Channel, error) {
	rows, err := db.QueryContext(ctx, listChannelsByProvider, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Channel{}
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ID,
			&i.ProviderID,
			&i.Provider,
			&i.Name,
			&i.Deleted,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listChannelsWithBots = `-- name: ListChannelsWithBots :many
WITH channelIds AS (
    SELECT distinct channel as id FROM bot_channel WHERE deleted IS NULL
)
SELECT id, provider_id, provider, name, deleted FROM channel WHERE id IN (SELECT id FROM channelIds) AND deleted IS NULL
`

func (q *Queries) ListChannelsWithBots(ctx context.Context, db DBTX) ([]*Channel, error) {
	rows, err := db.QueryContext(ctx, listChannelsWithBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Channel{}
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ID,
			&i.ProviderID,
			&i.Provider,
			&i.Name,
			&i.Deleted,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeBotChannel = `-- name: RemoveBotChannel :one
UPDATE bot_channel SET deleted = NOW() WHERE bot = $1 AND channel = $2 RETURNING bot
`

type RemoveBotChannelParams struct {
	Bot     uuid.UUID
	Channel uuid.UUID
}

func (q *Queries) RemoveBotChannel(ctx context.Context, db DBTX, arg RemoveBotChannelParams) (uuid.UUID, error) {
	row := db.QueryRowContext(ctx, removeBotChannel, arg.Bot, arg.Channel)
	var bot uuid.UUID
	err := row.Scan(&bot)
	return bot, err
}

const upsertBotChannel = `-- name: UpsertBotChannel :one
INSERT INTO bot_channel (bot, channel, provider) VALUES ($1, $2, $3) ON CONFLICT (bot, channel) DO UPDATE SET deleted = NULL RETURNING bot
`

type UpsertBotChannelParams struct {
	Bot      uuid.UUID
	Channel  uuid.UUID
	Provider Provider
}

func (q *Queries) UpsertBotChannel(ctx context.Context, db DBTX, arg UpsertBotChannelParams) (uuid.UUID, error) {
	row := db.QueryRowContext(ctx, upsertBotChannel, arg.Bot, arg.Channel, arg.Provider)
	var bot uuid.UUID
	err := row.Scan(&bot)
	return bot, err
}

const upsertChannel = `-- name: UpsertChannel :one
INSERT INTO channel (id, provider_id, provider, name)
SELECT coalesce(id, new_id), $1, $2, $3
FROM (VALUES(gen_random_uuid())) AS data(new_id) LEFT JOIN channel c
ON c.provider = $2 AND c.provider_id = $1
ON CONFLICT(provider_id, provider) DO UPDATE SET name = $3
RETURNING id, provider_id, provider, name, deleted
`

type UpsertChannelParams struct {
	ProviderID string
	Provider   Provider
	Name       string
}

func (q *Queries) UpsertChannel(ctx context.Context, db DBTX, arg UpsertChannelParams) (*Channel, error) {
	row := db.QueryRowContext(ctx, upsertChannel, arg.ProviderID, arg.Provider, arg.Name)
	var i Channel
	err := row.Scan(
		&i.ID,
		&i.ProviderID,
		&i.Provider,
		&i.Name,
		&i.Deleted,
	)
	return &i, err
}
