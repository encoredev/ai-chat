// Gemini is an llm provider implementation for the Google Gemini API.
package gemini

import (
	"context"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/cockroachdb/errors"
	"google.golang.org/api/option"

	"encore.app/llm/provider"
	"encore.app/pkg/fns"
	"encore.dev/config"
)

type Config struct {
	Project     config.String
	Model       config.String
	Region      config.String
	Temperature config.Float32
	TopK        config.Int32
}

var cfg = config.Load[*Config]()

var secrets struct {
	GeminiJSONCredentials string
}

//encore:service
type Service struct {
	client *genai.GenerativeModel
}

func initService() (*Service, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, cfg.Project(), "us-east1", option.WithCredentialsJSON([]byte(secrets.GeminiJSONCredentials)))
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	model := client.GenerativeModel(cfg.Model())
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
	}
	model.SetTemperature(cfg.Temperature())
	model.CandidateCount = fns.Ptr[int32](1)
	model.SetTopK(cfg.TopK())
	svc := &Service{
		client: model,
	}
	return svc, nil
}

type GenerateAvatarRequest struct {
	Prompt string
}

type GenerateAvatarResponse struct {
	Image []byte
}

//encore:api private method=POST path=/gemini/generate-avatar
func (p *Service) GenerateAvatar(ctx context.Context, req *GenerateAvatarRequest) (*GenerateAvatarResponse, error) {
	return nil, nil
}

type AskRequest struct {
	Message string
}

type AskResponse struct {
	Message string
}

func flattenResponse(resp *genai.GenerateContentResponse) string {
	var rtn strings.Builder
	for i, part := range resp.Candidates[0].Content.Parts {
		switch part := part.(type) {
		case genai.Text:
			if i > 0 {
				rtn.WriteString(" ")
			}
			rtn.WriteString(string(part))
		}
	}
	return rtn.String()
}

//encore:api private method=POST path=/gemini/ask
func (p *Service) Ask(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	session := p.client.StartChat()
	resp, err := session.SendMessage(ctx, genai.Text(req.Message))
	if err != nil {
		return nil, errors.Wrap(err, "send message")
	}
	return &AskResponse{Message: flattenResponse(resp)}, nil
}

type ContinueChatResponse struct {
	Message string
}

//encore:api private method=POST path=/gemini/continue-chat
func (p *Service) ContinueChat(ctx context.Context, req *provider.ChatRequest) (*ContinueChatResponse, error) {
	var history []*genai.Content
	var curMsg *genai.Content
	if req.SystemMsg != "" {
		curMsg = &genai.Content{
			Role: "user",
			Parts: []genai.Part{
				genai.Text(req.SystemMsg),
			},
		}
	}
	botNames := make(map[string]struct{})
	for _, b := range req.Bots {
		botNames[b.Name] = struct{}{}
	}
	for _, m := range req.Messages {
		role := "user"
		if req.FromBot(m) {
			role = "model"
		}
		if curMsg == nil || curMsg.Role != role {
			if curMsg != nil {
				history = append(history, curMsg)
			}
			curMsg = &genai.Content{
				Role:  role,
				Parts: []genai.Part{},
			}
		}
		curMsg.Parts = append(curMsg.Parts, genai.Text(req.Format(m)))
	}
	session := p.client.StartChat()
	session.History = history
	resp, err := session.SendMessage(ctx, curMsg.Parts...)
	if err != nil {
		return nil, errors.Wrap(err, "send message")
	}
	return &ContinueChatResponse{Message: flattenResponse(resp)}, nil
}
