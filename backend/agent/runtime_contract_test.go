package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/insmtx/Leros/backend/agent"
	clauderuntime "github.com/insmtx/Leros/backend/agent/runtime/claude"
	codexruntime "github.com/insmtx/Leros/backend/agent/runtime/codex"
	"github.com/insmtx/Leros/backend/agent/runtime/externalcli"
	nativeruntime "github.com/insmtx/Leros/backend/agent/runtime/native"
	opencoderuntime "github.com/insmtx/Leros/backend/agent/runtime/opencode"
	"github.com/insmtx/Leros/backend/agent/runtime/provider"
)

type contractEngine struct {
	request provider.RunRequest
	err     error
}

func (*contractEngine) Prepare(context.Context, provider.PrepareRequest) error { return nil }
func (*contractEngine) GetSkillDir() string                                    { return "" }

func (e *contractEngine) Run(
	_ context.Context,
	request provider.RunRequest,
) (*provider.RunHandle, error) {
	e.request = request
	if e.err != nil {
		return nil, e.err
	}
	events := make(chan agent.Event, 2)
	events <- agent.Event{Type: provider.EngineEventResult, Content: "done"}
	events <- agent.Event{Type: provider.EngineEventCompleted}
	close(events)
	return &provider.RunHandle{Events: events}, nil
}

func TestConcreteRuntimesFollowRuntimeContract(t *testing.T) {
	tests := []struct {
		name string
		wrap func(*externalcli.Driver) agent.Runtime
	}{
		{name: clauderuntime.Kind, wrap: func(driver *externalcli.Driver) agent.Runtime {
			return clauderuntime.New(driver)
		}},
		{name: codexruntime.Kind, wrap: func(driver *externalcli.Driver) agent.Runtime {
			return codexruntime.New(driver)
		}},
		{name: opencoderuntime.Kind, wrap: func(driver *externalcli.Driver) agent.Runtime {
			return opencoderuntime.New(driver)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := &contractEngine{}
			driver, err := externalcli.NewDriver(test.name, engine)
			if err != nil {
				t.Fatalf("NewDriver() error = %v", err)
			}
			driver.SetSessionStore(externalcli.NewInMemoryProviderSessionStore())
			runtime := test.wrap(driver)
			request := agent.ExecutionRequest{ExecutionID: "execution-1", Runtime: test.name}
			result, err := runtime.Execute(context.Background(), request, nil)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if runtime.Name() != test.name || result.Message != "done" {
				t.Fatalf("runtime name/result = %q/%q", runtime.Name(), result.Message)
			}
			if engine.request.ExecutionID != request.ExecutionID {
				t.Fatalf("forwarded request = %#v", engine.request)
			}

			engine.err = context.Canceled
			if _, err := runtime.Execute(context.Background(), request, nil); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancel error = %v", err)
			}
		})
	}
}

func TestNativeRuntimeFollowsRuntimeContract(t *testing.T) {
	runtime, err := nativeruntime.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if runtime.Name() != nativeruntime.Kind {
		t.Fatalf("Name() = %q", runtime.Name())
	}
	if _, err := runtime.Execute(context.Background(), agent.ExecutionRequest{}, nil); err == nil {
		t.Fatal("Execute() should reject an empty execution id")
	}
}
