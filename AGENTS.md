# AGENT DEVELOPMENT GUIDELINES FOR SINGEROS

This document contains essential information for AI agents working with the SingerOS codebase.

## BUILD/LINT/TEST COMMANDS

### Build Commands
- `go build -o ./bundles/singeros ./cmd/singeros/main.go` - Build the main SingerOS binary
- `go install ./cmd/singeros/main.go` - Install the SingerOS binary
- `make docker-build` - Build Docker image (tag: registry.yygu.cn/insmtx/SingerOS:latest)
- `make docker-run` - Run the Docker image locally

### Test Commands
- `go test ./...` - Run all tests in the project
- `go test -v ./...` - Run all tests with verbose output
- `go test ./pkg/path/to/package` - Run tests for a specific package
- `go test -run ^TestFunctionName$ ./pkg/path` - Run a specific test function
- `go test -race ./...` - Run all tests with race condition detection
- `go test -cover ./...` - Run tests and display coverage information

### Alternative Test Commands (from CONTRIBUTING.md)
- `make test` - Run all tests (as referenced in documentation)
- `make test-cover` - Run tests with coverage (as referenced in documentation)

### Lint Commands
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Vet all Go code for common mistakes
- `golint ./...` - Lint all Go code (install via `go install golang.org/x/lint/golint@latest`)
- `gofmt -s -w .` - Simplify code and write changes (as per the existing Makefile)
- `staticcheck ./...` - Comprehensive Go static analysis (if installed)

## CODE STYLE GUIDELINES

### Import Organization
- Group imports with blank lines between standard library, third-party, and project-specific packages
- Use semantic import aliases only when they prevent naming conflicts
- Organize in three groups: stdlib, third-party, internal packages
```
import (
	"fmt"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	
	"github.com/insmtx/SingerOS/internal/config"
)
```

### Formatting Conventions
- Use tabs for indentation, not spaces (as verified from existing Go files)
- Execute `go fmt ./...` before committing
- Keep lines under 120 characters where possible
- Use `gofmt -s` for simplification of code

### Naming Conventions
- Use CamelCase for exported functions/types (`GetUser`, `UserService`)
- Use camelCase for unexported/internal functions/types (`getUser`, `userService`)
- Use clear, descriptive names; prefer clarity over brevity
- Use consistent names for similar concepts across packages
- Variables related to the system should reference SingerOS concepts

### Types and Interfaces
- Define interfaces close to their first usage
- Keep interfaces small, typically one or a few methods
- Name interface types with "-er" suffix when applicable (e.g., `Runner`, `Handler`)
- Use concrete types explicitly in function signatures when interface is not needed
- Prefer returning pointers for structs when passing to functions if they will be modified

### Error Handling
- Handle errors explicitly; don't ignore them
- Use specific error types when appropriate with wrapped errors
- Follow the pattern: "if err != nil { return err }"
- Use `errors.New()` for simple static strings
- Use `fmt.Errorf()` with `%w` verb for wrapping errors with more context
- Log errors contextually when appropriate

### Additional Guidelines
- All public functions must have GoDoc comments
- Comments should be in English and explain why rather than what
- Maintain consistent logging format throughout the application
- Use context.Context appropriately for cancellation and request-scoped values
- Follow dependency injection patterns rather than global variables
- Use Cobra for command-line interface implementations as shown in main.go files

### Commit Guidelines
- Follow conventional commits format: `<type>(<scope>): <subject>`
- Use Chinese for commit messages in SingerOS project
- Type options include:
  - `feat`: New feature
  - `fix`: Bug fixes
  - `docs`: Documentation updates
  - `style`: Code style adjustments
  - `refactor`: Code refactoring
  - `test`: Testing related
  - `chore`: Build tool or auxiliary tool changes
- When applicable, include detailed descriptions in the body covering technical implementation and business logic

## PROJECT STRUCTURE

- `/backend` - Main Go application code
- `/backend/cmd` - Entry points for different SingerOS services
- `/internal` - Private internal code that should not be imported by other projects
- `/pkg` - Public libraries that can be used by other applications
- `/docs` - Documentation files
- `/deployments/build/Dockerfile` - Container build configuration

## CONTRIBUTION NOTES

- See CONTRIBUTING.md for commit message style guidance
- Make sure all tests pass (`go test ./...`) before submitting changes
- Follow Go's idiomatic patterns and standard practices
- When implementing, consider how components fit into the broader microservices architecture described in ARCHITECTURE.md

## CORE COMPONENTS AND ARCHITECTURE

Based on the AI OS architecture described in docs/ARCHITECTURE_v2.md, the SingerOS platform consists of the following primary components:

1. **Event Gateway** - Receives external events from various channels
2. **Event Bus** - Message queue system for decoupling components
3. **Orchestrator** - Core scheduling and coordination mechanism
4. **DigitalAssistant** - Top-level abstraction representing AI workers
5. **Agent** - Decision-making entities within DigitalAssistants
6. **Skill** - Reusable capabilities that can be invoked
7. **Tool** - External system integrations that Skills can call

## SKILL SYSTEM DEFINITION

Skills represent core building blocks in SingerOS:

```
skill.id
skill.name
skill.description
skill.input_schema
skill.output_schema
skill.permissions
skill.executor
```

### Skill Categories

- **Integration Skills** - External system integrations (GitHub, GitLab, WeChat, Feishu, Jira)
- **AI Skills** - LLM-based reasoning capabilities (code_review, summarize, classification)
- **Tool Skills** - Utility capabilities (run_shell, execute_python, http_request)
- **Workflow Skills** - Complex coordinated operations (pr_review_workflow, bug_triage_workflow)

## CHANNEL INTEGRATION

Support for multiple interaction channels:

- GitHub
- GitLab
- Enterprise WeChat
- Feishu
- App
- Webhook

Each channel is abstracted through a Channel adapter pattern for unified interaction handling.

## PERMISSIONS AND SECURITY

Granular permissions control at multiple levels:

- DigitalAssistant
- Agent
- Skill
- Tool

Permission model: RBAC + Capability

## GOLANG ENGINE STRUCTURE

SingerOS follows this recommended Golang project structure:

```
aios/
│
├── cmd/
│   ├── api
│   ├── worker
│   └── scheduler
│
├── internal/
│   ├── core/
│   │   ├── employee
│   │   ├── agent
│   │   ├── workflow
│   │   ├── skill
│   │   └── event
│
│   ├── orchestrator/
│   │   └── orchestrator.go
│
│   ├── engine/
│   │   ├── agent_engine
│   │   ├── workflow_engine
│   │   └── skill_engine
│
│   ├── integrations/
│   │   ├── github
│   │   ├── gitlab
│   │   ├── wechat
│   │   └── feishu
│
│   ├── skills/
│   │   ├── git
│   │   ├── opencode
│   │   ├── llm
│   │   └── messaging
│
│   ├── storage/
│   │   ├── postgres
│   │   ├── redis
│   │   └── vector
│
│   ├── eventbus/
│   │   └── nats
│
│   ├── auth/
│   │   └── permissions
│
│   └── config/
│
├── pkg/
│   ├── sdk
│   └── client
│
├── api/
│   └── proto
│
├── deployments/
│   └── docker
│
└── docs/
```

## MINIMUM VISION PRODUCT (MVP)

The initial MVP focuses on these key components:

1. Event Gateway
2. Orchestrator
3. Skill System
4. GitHub Integration
5. CodeAssistantEmployee

MVP Features:

- PR automatic Review
- PR automatic summary
- Issue automatic reply
- Code explanation

## TECHNICAL STACK

Recommended stack:

- Language: Golang
- Message System: NATS
- Database: Postgres
- Cache: Redis
- Vector Store: Qdrant
- LLM: OpenAI / Claude / DeepSeek
