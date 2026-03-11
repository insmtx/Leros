package eventbus

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, event any) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler func(event any)) error
}

type EventBus struct {
	publisher  Publisher
	subscriber Subscriber
}

func NewEventBus(publisher Publisher, subscriber Subscriber) *EventBus {
	return &EventBus{
		publisher:  publisher,
		subscriber: subscriber,
	}
}
