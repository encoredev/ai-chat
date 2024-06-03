package llm

import (
	"context"

	"github.com/cockroachdb/errors"

	"encore.app/llm/provider"
	"encore.dev/pubsub"
	"encore.dev/rlog"
)

type TaskType string

const (
	TaskTypeJoin        TaskType = "join"
	TaskTypeLeave       TaskType = "leave"
	TaskTypeContinue    TaskType = "continue"
	TaskTypeInstruct    TaskType = "instruct"
	TaskTypePrepopulate TaskType = "prepopulate"
)

type Task struct {
	Type    TaskType
	Request *provider.ChatRequest
}

// TaskTopic is a topic for tasks generated by the chat service.
// We use a topic to allow for asynchronous processing of tasks (the LLMs can be slow).
//
// This uses Encore's pubsub package, learn more: https://encore.dev/docs/primitives/pubsub
var TaskTopic = pubsub.NewTopic[*Task]("llm-task", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// chat-sub is a subscription to the chat task topic. It handles all incoming tasks from the chat service.
//
// This uses Encore's pubsub package, learn more: https://encore.dev/docs/primitives/pubsub
var _ = pubsub.NewSubscription(
	TaskTopic, "chat-sub",
	pubsub.SubscriptionConfig[*Task]{
		Handler: pubsub.MethodHandler((*Service).ProcessTask),
	},
)

// ProcessTask processes a task from the chat service by forwarding the request to the appropriate provider.
//
//encore:api private method=POST path=/ai/task
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
	case TaskTypePrepopulate:
		res, err = svc.Prepopulate(ctx, task.Request)
		if err != nil {
			return errors.Wrap(err, "prepopulate")
		}
	}
	if len(res.Messages) > 0 {
		_, err := LLMMessageTopic.Publish(ctx, res)
		if err != nil {
			rlog.Warn("publish message", "error", err)
		}
	}
	return nil
}

// LLMMessageTopic is a topic for messages generated by the LLM providers.
//
// This uses Encore's pubsub package, learn more: https://encore.dev/docs/primitives/pubsub
var LLMMessageTopic = pubsub.NewTopic[*BotResponse]("llm-messages", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
