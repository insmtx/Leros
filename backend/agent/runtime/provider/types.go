// Package provider defines provider process contracts used by concrete runtimes.
package provider

import "github.com/insmtx/Leros/backend/agent"

const (
	// EngineEventStarted indicates the engine process has started.
	EngineEventStarted agent.EventType = "engine.started"
	// EngineEventProviderSessionStarted indicates the provider exposed a native session ID.
	EngineEventProviderSessionStarted agent.EventType = "provider_session.started"
	// EngineEventMessageDelta contains assistant text output.
	EngineEventMessageDelta agent.EventType = "message.delta"
	// EngineEventReasoningDelta contains reasoning/thinking output.
	EngineEventReasoningDelta agent.EventType = "reasoning.delta"
	// EngineEventResult contains the final assistant result (before completion).
	EngineEventResult agent.EventType = "message.result"
	// EngineEventToolCallStarted indicates a tool call has started.
	EngineEventToolCallStarted agent.EventType = "tool_call.started"
	// EngineEventToolCallCompleted indicates a tool call has completed successfully.
	EngineEventToolCallCompleted agent.EventType = "tool_call.completed"
	// EngineEventToolCallFailed indicates a tool call has failed.
	EngineEventToolCallFailed agent.EventType = "tool_call.failed"
	// EngineEventTodoSnapshot contains the current full todo list.
	EngineEventTodoSnapshot agent.EventType = "todo.snapshot"
	// EngineEventTodoUpdated contains an updated full todo list.
	EngineEventTodoUpdated agent.EventType = "todo.updated"
	// EngineEventApprovalRequested indicates the engine needs user approval.
	EngineEventApprovalRequested agent.EventType = "approval.requested"
	// EngineEventApprovalResolved indicates an approval request was resolved.
	EngineEventApprovalResolved agent.EventType = "approval.resolved"
	// EngineEventQuestionAsked indicates the engine is asking a clarifying question.
	EngineEventQuestionAsked agent.EventType = "question.asked"
	// EngineEventQuestionAnswered indicates a question was answered.
	EngineEventQuestionAnswered agent.EventType = "question.answered"
	// EngineEventCompleted indicates the engine process completed successfully.
	EngineEventCompleted agent.EventType = "engine.completed"
	// EngineEventFailed indicates the engine process failed.
	EngineEventFailed agent.EventType = "engine.failed"
	// EngineEventCancelled indicates the engine process was cancelled.
	EngineEventCancelled agent.EventType = "engine.cancelled"
)
