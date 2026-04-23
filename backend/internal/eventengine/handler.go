package eventengine

import (
	"context"

	"github.com/insmtx/SingerOS/backend/pkg/event"
	"github.com/ygpkg/yg-go/logs"
)

func (e *EventEngine) registerDefaultHandlers() {
	e.handlers[event.TopicGithubIssueComment] = e.handleIssueComment
	e.handlers[event.TopicGithubPullRequest] = e.handlePullRequest
	e.handlers[event.TopicGithubPush] = e.handlePush
}

func (e *EventEngine) Start(ctx context.Context) error {
	for topic, handler := range e.handlers {
		go func(t string, h EventHandler) {
			logs.InfoContextf(ctx, "Starting subscription for topic: %s", t)
			err := e.subscriber.Subscribe(ctx, t, func(rawEvent any) {
				interactionEvent, ok := rawEvent.(*event.Event)
				if !ok {
					logs.ErrorContextf(ctx, "Invalid event type received")
					return
				}

				logs.DebugContextf(ctx, "Received event on topic %s: %+v", t, interactionEvent)

				if err := h(ctx, interactionEvent); err != nil {
					logs.ErrorContextf(ctx, "Error handling event on topic %s: %v", t, err)
				}
			})

			if err != nil {
				logs.ErrorContextf(ctx, "Failed to subscribe to topic %s: %v", t, err)
			}
		}(topic, handler)
	}

	return nil
}

func (e *EventEngine) RegisterHandler(topic string, handler EventHandler) {
	e.handlers[topic] = handler
}

func (e *EventEngine) GetHandler(topic string) (EventHandler, error) {
	handler, exists := e.handlers[topic]
	if !exists {
		return nil, nil
	}
	return handler, nil
}

func (e *EventEngine) handleIssueComment(ctx context.Context, evt *event.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub issue comment event: %+v", evt)
	return nil
}

func (e *EventEngine) handlePullRequest(ctx context.Context, evt *event.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub pull request event: %+v", evt)
	return nil
}

func (e *EventEngine) handlePush(ctx context.Context, evt *event.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub push event: %+v", evt)
	return nil
}
