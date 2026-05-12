package taskconsumer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/insmtx/Leros/backend/runtime/events"
	"github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/pkg/dm"
)

func TestPublishWorkerTaskMessageToNATS(t *testing.T) {
	natsURL := getenv("LEROS_TEST_NATS_URL", "nats://localhost:4222")
	orgID := getenvUint("LEROS_TEST_ORG_ID", 1001)
	workerID := getenvUint("LEROS_TEST_WORKER_ID", 1)
	sessionID := getenv("LEROS_TEST_SESSION_ID", "session_1")

	bus, err := mq.NewPublisher(natsURL)
	if err != nil {
		t.Skipf("skip real NATS publish test: %v", err)
	}
	defer bus.Close()

	topic, _ := dm.WorkerTaskTopic(orgID, workerID)
	messageID := randomTestID(t, "msg")
	traceID := randomTestID(t, "trace")
	requestID := randomTestID(t, "request")
	taskID := randomTestID(t, "task")
	runID := randomTestID(t, "run")

	msg := events.WorkerTaskMessage{
		ID:        messageID,
		Type:      events.MessageTypeWorkerTask,
		CreatedAt: time.Now().UTC(),
		Trace: events.TraceContext{
			TraceID:   traceID,
			RequestID: requestID,
			TaskID:    taskID,
			RunID:     runID,
		},
		Route: events.RouteContext{
			OrgID:     orgID,
			SessionID: sessionID,
			WorkerID:  workerID,
		},
		Body: events.WorkerTaskBody{
			TaskType: events.TaskTypeAgentRun,
			Actor: events.ActorContext{
				UserID:      "user_test",
				DisplayName: "Test User",
				Channel:     "go_test",
			},
			Execution: events.ExecutionTarget{
				AssistantID: "assistant_test",
				AgentID:     "agent_test",
				Tools:       []string{},
			},
			Input: events.TaskInput{
				Type: events.InputTypeTaskInstruction,
				Text: "这是一条来自 go test 的真实 NATS worker.task 调试消息，请回复确认 worker 已收到。",
			},
			Runtime: events.RuntimeOptions{
				Kind:    "claude",
				WorkDir: ".",
			},
		},
		Metadata: map[string]any{
			"source": "go_test",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := bus.Publish(ctx, topic, msg); err != nil {
		t.Fatalf("Publish(%q) error = %v", topic, err)
	}
	t.Logf(
		"published worker task:\n  topic: %s\n  nats_url: %s\n  message_id: %s\n  trace_id: %s\n  request_id: %s\n  task_id: %s\n  run_id: %s",
		topic,
		natsURL,
		messageID,
		traceID,
		requestID,
		taskID,
		runID,
	)
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvUint(key string, fallback uint) uint {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}
	value, err := strconv.ParseUint(valueStr, 10, 32)
	if err != nil {
		return fallback
	}
	return uint(value)
}

func randomTestID(t *testing.T, prefix string) string {
	t.Helper()

	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		t.Fatalf("generate %s id: %v", prefix, err)
	}
	return fmt.Sprintf("%s_test_agent_run_%s", prefix, hex.EncodeToString(buf[:]))
}
