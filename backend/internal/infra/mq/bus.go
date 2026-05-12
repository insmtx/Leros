// mq 包提供消息队列的抽象和实现
//
// 该包定义了事件发布者和订阅者的标准接口，以及基于 RabbitMQ 的实现。
// 支持多种事件总线实现，如 RabbitMQ、Redis 等。
package mq

import (
	"context"
	"github.com/nats-io/nats.go"
)

// Publisher 是事件发布者接口，定义了向指定主题发布事件的方法
type Publisher interface {
	// Publish 向指定主题发布事件
	Publish(ctx context.Context, topic string, event any) error
}

// RealtimePublisher 是实时消息发布者接口，不要求 MQ 层持久化。
type RealtimePublisher interface {
	// PublishRealtime 发布实时消息，持久化由接收端按需处理。
	PublishRealtime(ctx context.Context, topic string, event any) error
}

// Subscriber 是事件订阅者接口，定义了订阅指定主题事件的方法
type Subscriber interface {
	// Subscribe 订阅指定主题的事件，并使用提供的处理函数处理收到的事件
	Subscribe(ctx context.Context, topic string, handler func(msg *nats.Msg)) error
}

// RealtimeSubscriber 是实时消息订阅者接口，用于接收 PublishRealtime 发布的消息。
type RealtimeSubscriber interface {
	// SubscribeRealtime 订阅实时消息主题，不依赖 JetStream 持久化。
	SubscribeRealtime(ctx context.Context, topic string, handler func(msg *nats.Msg)) error
}

// EventBus 组合了发布和订阅能力，提供完整的事件总线功能
type EventBus interface {
	Publisher
	Subscriber
	RealtimePublisher
	RealtimeSubscriber
}
