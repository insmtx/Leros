// Package eventengine subscribes to interaction events and dispatches agent runs.
package eventengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/insmtx/Leros/backend/internal/agent"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	interactionevent "github.com/insmtx/Leros/backend/pkg/event"
	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

// EventHandlerFunc handles one normalized interaction event.
type EventHandlerFunc func(ctx context.Context, event *interactionevent.Event) error

// Orchestrator subscribes to interaction event topics and dispatches them to the agent runtime.
type Orchestrator struct {
	subscriber eventbus.Subscriber
	runner     agent.Runner
	handlers   map[string]EventHandlerFunc
}

// NewOrchestrator creates a new event orchestrator.
func NewOrchestrator(subscriber eventbus.Subscriber, runner agent.Runner) *Orchestrator {
	orchestrator := &Orchestrator{
		subscriber: subscriber,
		runner:     runner,
		handlers:   make(map[string]EventHandlerFunc),
	}
	orchestrator.registerDefaultHandlers()
	return orchestrator
}

func (o *Orchestrator) registerDefaultHandlers() {
	o.handlers[interactionevent.TopicGithubIssueComment] = o.handleIssueComment
	o.handlers[interactionevent.TopicGithubPullRequest] = o.handlePullRequest
	o.handlers[interactionevent.TopicGithubPush] = o.handlePush
}

// Start starts subscriptions for registered event topics.
func (o *Orchestrator) Start(ctx context.Context) error {
	for topic, handler := range o.handlers {
		go func(t string, h EventHandlerFunc) {
			logs.InfoContextf(ctx, "Starting subscription for topic: %s", t)
			err := o.subscriber.SubscribeFrom(ctx, t, 0, func(msg *nats.Msg) {
				var interactionEvent interactionevent.Event
				if err := json.Unmarshal(msg.Data, &interactionEvent); err != nil {
					logs.ErrorContextf(ctx, "Failed to unmarshal event: %v", err)
					return
				}

				logs.DebugContextf(ctx, "Received event on topic %s: %+v", t, interactionEvent)
				if err := h(ctx, &interactionEvent); err != nil {
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

func (o *Orchestrator) handleIssueComment(ctx context.Context, event *interactionevent.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub issue comment event with agent runtime: %+v", event)
	return o.runEvent(ctx, event)
}

func (o *Orchestrator) handlePullRequest(ctx context.Context, event *interactionevent.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub pull request event with agent runtime: %+v", event)
	return o.runEvent(ctx, event)
}

func (o *Orchestrator) handlePush(ctx context.Context, event *interactionevent.Event) error {
	logs.InfoContextf(ctx, "Processing GitHub push event with agent runtime: %+v", event)
	return o.runEvent(ctx, event)
}

func (o *Orchestrator) runEvent(ctx context.Context, event *interactionevent.Event) error {
	if o.runner == nil {
		return fmt.Errorf("agent runtime runner is required")
	}
	result, err := o.runner.Run(ctx, requestFromInteractionEvent(event))
	if err != nil {
		return err
	}
	if result != nil {
		logs.InfoContextf(ctx, "Agent runtime completed event: run_id=%s status=%s", result.RunID, result.Status)
	}
	return nil
}

// RegisterHandler registers an event handler for a topic.
func (o *Orchestrator) RegisterHandler(topic string, handler EventHandlerFunc) {
	o.handlers[topic] = handler
}

// GetHandler returns the handler registered for a topic.
func (o *Orchestrator) GetHandler(topic string) (EventHandlerFunc, error) {
	handler, exists := o.handlers[topic]
	if !exists {
		return nil, fmt.Errorf("no handler registered for topic: %s", topic)
	}
	return handler, nil
}
