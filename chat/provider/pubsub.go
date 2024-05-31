package provider

import (
	"encore.app/chat/service/clients"
	"encore.dev/pubsub"
)

var MessageTopic = pubsub.NewTopic[*client.Message]("provider-messages", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
