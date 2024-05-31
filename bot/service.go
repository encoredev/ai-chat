// The bot service is responsible for managing bots. It allows creating, listing, getting, and deleting bots.
// It also provides an endpoint to get the avatar of a bot which is exposed as a raw endpoint.
package bot

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"encore.app/bot/db"
	"encore.app/llm/service"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

type BotID = uuid.UUID

var botdb = sqldb.NewDatabase("bot", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})

//encore:service
type Service struct{}

type CreateBotRequest struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	LLM    string `json:"llm"`
}

//encore:api public method=POST path=/bots
func (svc *Service) Create(ctx context.Context, req *CreateBotRequest) (*db.Bot, error) {
	resp, err := llm.GenerateBotProfile(ctx, &llm.GenerateBotProfileRequest{
		Name:     req.Name,
		Prompt:   req.Prompt,
		Provider: req.LLM,
	})
	if err != nil {
		return nil, err
	}
	return db.New().InsertBot(ctx, botdb.Stdlib(), db.InsertBotParams{
		Name:     req.Name,
		Profile:  resp.Profile,
		Prompt:   req.Prompt,
		Avatar:   resp.Avatar,
		Provider: req.LLM,
	})
}

type Bots struct {
	Bots []*db.Bot `json:"bots"`
}

type ListBotRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

//encore:api public method=GET path=/bots
func (svc *Service) List(ctx context.Context, req *ListBotRequest) (*Bots, error) {
	bots, err := func() ([]*db.Bot, error) {
		if len(req.IDs) == 0 {
			return db.New().ListBot(ctx, botdb.Stdlib())
		}
		return db.New().GetBots(ctx, botdb.Stdlib(), req.IDs)
	}()
	if err != nil {
		return nil, err
	}
	return &Bots{Bots: bots}, nil
}

//encore:api public method=GET path=/bots/:id
func (svc *Service) Get(ctx context.Context, id uuid.UUID) (*db.Bot, error) {
	return db.New().GetBot(ctx, botdb.Stdlib(), id)
}

//encore:api public method=DELETE path=/bots/:id
func (svc *Service) Delete(ctx context.Context, id uuid.UUID) (*db.Bot, error) {
	return db.New().DeleteBot(ctx, botdb.Stdlib(), id)
}

//encore:api public raw path=/bots/:id/avatar
func (*Service) Avatar(w http.ResponseWriter, req *http.Request) {
	q := db.New()
	id := strings.TrimPrefix(req.RequestURI, "/bots/")
	id = strings.TrimSuffix(id, "/avatar")
	uid, err := uuid.FromString(id)
	if err != nil {
		http.Error(w, "Invalid ProviderID", http.StatusBadRequest)
		return
	}
	bot, err := q.GetBot(req.Context(), botdb.Stdlib(), uid)
	if err != nil {
		http.Error(w, "Bot not found", http.StatusNotFound)
		return
	}
	if bot.Avatar == nil {
		http.Error(w, "Bot has no avatar", http.StatusNotFound)
		return
	}
	http.ServeContent(w, req, "avatar.png", time.Now(), bytes.NewReader(bot.Avatar))
}
