// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"encore.dev/types/uuid"
)

type Provider string

const (
	ProviderSlack      Provider = "slack"
	ProviderDiscord    Provider = "discord"
	ProviderAdmin      Provider = "admin"
	ProviderEncorechat Provider = "localchat"
)

func (e *Provider) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Provider(s)
	case string:
		*e = Provider(s)
	default:
		return fmt.Errorf("unsupported scan type for Provider: %T", src)
	}
	return nil
}

type NullProvider struct {
	Provider Provider
	Valid    bool // Valid is true if Provider is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullProvider) Scan(value interface{}) error {
	if value == nil {
		ns.Provider, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Provider.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullProvider) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Provider), nil
}

type BotChannel struct {
	Bot      uuid.UUID
	Channel  uuid.UUID
	Provider Provider
	Deleted  sql.NullTime
}

type Channel struct {
	ID         uuid.UUID
	ProviderID string
	Provider   Provider
	Name       string
	Deleted    sql.NullTime
}

type Message struct {
	ID         uuid.UUID
	ProviderID string
	ChannelID  uuid.UUID
	AuthorID   uuid.UUID
	Content    string
	Timestamp  time.Time
	Deleted    sql.NullTime
}

type User struct {
	ID         uuid.UUID
	Provider   Provider
	ProviderID string
	Name       string
	Profile    string
	BotID      *uuid.UUID
}
