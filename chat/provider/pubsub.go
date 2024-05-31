package provider

import (
	"encore.app/chat/service/client"
	"encore.dev/pubsub"
)

// MessageTopic is the pubsub topic for messages from chat providers
//
// This uses Encore's pubsub package, learn more: https://encore.dev/docs/primitives/pubsub
var MessageTopic = pubsub.NewTopic[*client.Message]("provider-messages", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
