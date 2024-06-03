package provider

import (
	"encore.app/chat/service/client"
	"encore.dev/pubsub"
)

// InboxTopic is the pubsub topic for messages from chat providers
//
// This uses Encore's pubsub package, learn more: https://encore.dev/docs/primitives/pubsub
var InboxTopic = pubsub.NewTopic[*client.Message]("inbox", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
