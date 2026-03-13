# Interaction Module

The Interaction module serves as the **Event Gateway** in the SingerOS AI OS architecture. It handles all external communication and event processing for the system, supporting multiple channels like GitHub, GitLab, WeChat, Feishu and others.

## Overview

The interaction module implements the event-driven architecture pattern described in the overall system design. It receives events from various external systems and routes them into the internal event bus for further processing by agents and workflows.

## Architecture Context

As per the SingerOS architecture:

```
External Systems (GitHub/GitLab/WeChat/Feishu/App/Webhook)
    ↓
Event Gateway (interaction module)
    ↓
Internal Event via Event Bus
    ↓
Orchestrator
    ↓
Agents → Skills → Workflows
```

## Components

### 1. Event Structure

The module defines a standard event format that abstracts external inputs into internal events:

- `EventID` - Unique identifier for the event
- `TraceID` - Traceability across the system  
- `Channel` - Source channel (github, gitlab, email, etc.)
- `EventType` - Type of incoming event
- `Actor` - User who initiated the event
- `Repository` - Repository context if applicable
- `Context` - Additional contextual information
- `Payload` - Original payload data
- `CreatedAt` - Timestamp of event creation

### 2. Channel Connectors

The module supports multiple interaction channels through a connector pattern:

- **GitHub Connector** - Handles GitHub webhooks and API
- **GitLab Connector** - Handles GitLab webhooks and API  
- **WeChat Connector** - Handles enterprise WeChat messages
- **Feishu Connector** - Handles Feishu (Lark) messages
- **General API Channel** - REST endpoint interactions

Connectors implement the `Connector` interface with:
- `ChannelCode()` - Returns a unique channel identifier
- `RegisterRoutes()` - Sets up webhooks/API routes

### 3. Event Processing Flow

Incoming external events are:
1. Received through channel-specific handlers
2. Converted to standardized internal `Event` structures
3. Published to the event bus with predefined topics
4. Routed internally by the orchestrator for processing

Current example topic:
- `interaction.github.issue_comment` - For GitHub issue comments

### 4. Event Gateway

The `gateway/event_gateway.go` file orchestrates the entire process by:
- Setting up routing based on configuration
- Registering all enabled connectors
- Publishing standardized events to the internal bus

## Topics Definition

Well-known event topics are defined in `topic.go` to ensure consistent routing of events throughout the system. These follow the convention `interaction.<channel>.<specific_event>`.

## Usage

The interaction module is typically initialized at application startup, where:

```go
registry := interaction.NewRegistry()
// Register all enabled connectors
registry.Register(githubConnector)
registry.Register(gitlabConnector)
registry.Register(weworkConnector)
// Set up routes in the main gin router
registry.RegisterRoutes(mainRouter)
```

## Security & Permissions

Interactions connect with appropriate authentication based on each channel:
- OAuth/JWT tokens for platforms like GitHub and GitLab 
- Secret verification for webhook endpoints
- Internal authorization for API endpoints

The system follows multi-layer permission control (DigitalAssistant → Agent → Skill → Tool) upon receiving interaction events.

## Current Implementation Status

Currently the system has partial GitHub integration implemented, with basic webhook handling in place. The foundation is established for adding additional channel integrations following the same pattern. All external events are normalized to the common `Event` structure for consistent internal handling.