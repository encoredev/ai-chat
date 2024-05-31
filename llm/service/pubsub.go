package llm

import (
	"encore.app/llm/provider"
	"encore.dev/pubsub"
)

type TaskType string

const (
	TaskTypeJoin     TaskType = "join"
	TaskTypeLeave    TaskType = "leave"
	TaskTypeContinue TaskType = "continue"
	TaskTypeInstruct TaskType = "instruct"
)

type Task struct {
	Type    TaskType
	Request *provider.ChatRequest
}

var TaskTopic = pubsub.NewTopic[*Task]("llm-task", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var _ = pubsub.NewSubscription(
	TaskTopic, "chat-sub",
	pubsub.SubscriptionConfig[*Task]{
		Handler: pubsub.MethodHandler((*Service).ProcessTask),
	},
)

var BotMessageTopic = pubsub.NewTopic[*BotResponse]("bot-messages", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
