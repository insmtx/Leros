package taskconsumer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/dm"
)

// const AgentRuntime = "leros"

const AgentRuntime = "claude"

// const AgentRuntime = "codex"

func TestPublishWorkerTaskMessageToNATS(t *testing.T) {
	natsURL := getenv("LEROS_TEST_NATS_URL", "nats://localhost:4222")
	orgID := getenvUint("LEROS_TEST_ORG_ID", 1)
	workerID := getenvUint("LEROS_TEST_WORKER_ID", 1)
	sessionID := getenv("LEROS_TEST_SESSION_ID", "session_1")
	healthURL := getenv("LEROS_TEST_WORKER_HEALTH_URL", "http://127.0.0.1:8081/health")

	checkWorkerHealthOrSkip(t, healthURL, orgID, workerID)

	bus, err := mq.NewNATS(natsURL)
	if err != nil {
		t.Skipf("skip real NATS publish test: %v", err)
	}
	defer bus.Close()

	topic, _ := dm.WorkerTaskSubject(orgID, workerID)
	streamTopic, _ := dm.SessionResultStreamSubject(orgID, sessionID)
	completedTopic, _ := dm.SessionMessageCompletedSubject(orgID, sessionID)

	task := newTestWorkerTaskMessage(t, orgID, workerID, sessionID)

	ctx, cancel := context.WithTimeout(context.Background(), getenvDuration("LEROS_TEST_AGENT_RUN_TIMEOUT", 2*time.Minute))
	defer cancel()

	receiveReady := make(chan error, 1)
	receiveDone := make(chan error, 1)
	go func() {
		receiveDone <- receiveWorkerTaskReply(ctx, t, bus, streamTopic, completedTopic, task.Trace.TaskID, task.Trace.RunID, receiveReady)
	}()

	if err := <-receiveReady; err != nil {
		t.Skipf("skip real NATS publish test: subscribe reply topics completed=%s: %v", completedTopic, err)
	}

	sendWorkerTaskMessage(ctx, t, bus, natsURL, topic, task)

	if err := <-receiveDone; err != nil {
		t.Fatal(err)
	}
}

// newTestWorkerTaskMessage builds a worker task message for integration tests.
func newTestWorkerTaskMessage(t *testing.T, orgID uint, workerID uint, sessionID string) protocol.WorkerTaskMessage {
	t.Helper()

	return protocol.WorkerTaskMessage{
		ID:        randomTestID(t, "msg"),
		Type:      protocol.MessageTypeWorkerTask,
		CreatedAt: time.Now().UTC(),
		Trace: protocol.TraceContext{
			TraceID:   randomTestID(t, "trace"),
			RequestID: randomTestID(t, "request"),
			TaskID:    randomTestID(t, "task"),
			RunID:     randomTestID(t, "run"),
		},
		Route: protocol.RouteContext{
			OrgID:     orgID,
			SessionID: sessionID,
			WorkerID:  workerID,
		},
		Body: protocol.WorkerTaskBody{
			TaskType: protocol.TaskTypeAgentRun,
			Actor: protocol.ActorContext{
				UserID:      "user_test",
				DisplayName: "Test User",
				Channel:     "go_test",
			},
			Execution: protocol.ExecutionTarget{
				AssistantID: "assistant_test",
				AgentID:     "agent_test",
				Tools:       []string{},
			},
			Input: protocol.TaskInput{
				Type: protocol.InputTypeTaskInstruction,
				Text: "Check the current system time and report it back.",
			},
			Runtime: protocol.RuntimeOptions{
				Kind:    AgentRuntime,
				WorkDir: ".",
			},
			Model: protocol.ModelOptions{
				ID: 1,
			},
		},
		Metadata: map[string]any{
			"source": "go_test",
		},
	}
}

// sendWorkerTaskMessage publishes a worker task message to the task topic.
func sendWorkerTaskMessage(ctx context.Context, t *testing.T, publisher mq.Publisher, natsURL string, topic string, msg protocol.WorkerTaskMessage) {
	t.Helper()

	if err := publisher.Publish(ctx, topic, msg); err != nil {
		t.Fatalf("Publish(%q) error = %v", topic, err)
	}
	t.Logf(
		"published worker task:\n  topic: %s\n  nats_url: %s\n  message_id: %s\n  trace_id: %s\n  request_id: %s\n  task_id: %s\n  run_id: %s",
		topic,
		natsURL,
		msg.ID,
		msg.Trace.TraceID,
		msg.Trace.RequestID,
		msg.Trace.TaskID,
		msg.Trace.RunID,
	)
}

// receiveWorkerTaskReply waits for the current task completion message.
func receiveWorkerTaskReply(ctx context.Context, t *testing.T, subscriber mq.Subscriber, streamTopic string, completedTopic string, taskID string, runID string, ready chan<- error) error {
	t.Helper()

	completedCh := make(chan protocol.MessageStreamMessage, 1)

	go func() {
		ready <- nil
		err := subscriber.SubscribeFrom(ctx, streamTopic, 0, func(natsMsg *nats.Msg) {
			var streamMsg protocol.MessageStreamMessage
			if err := json.Unmarshal(natsMsg.Data, &streamMsg); err != nil {
				t.Logf("topic %s malformed: %v\n%s", streamTopic, err, string(natsMsg.Data))
				return
			}
			t.Logf("topic %s event=%s content=%s\n%s", streamTopic, streamMsg.Body.Event, streamMsg.Body.Payload.Content, string(natsMsg.Data))
		})
		if err != nil {
			t.Logf("stream topic subscription error: %v", err)
		}
	}()

	go func() {
		ready <- nil
		err := subscriber.SubscribeFrom(ctx, completedTopic, 0, func(natsMsg *nats.Msg) {
			var completedMsg protocol.MessageStreamMessage
			if err := json.Unmarshal(natsMsg.Data, &completedMsg); err != nil {
				t.Logf("topic %s malformed: %v\n%s", completedTopic, err, string(natsMsg.Data))
				return
			}
			t.Logf("topic %s event=%s content=%s\n%s", completedTopic, completedMsg.Body.Event, completedMsg.Body.Payload.Content, string(natsMsg.Data))

			if completedMsg.Trace.TaskID != taskID && completedMsg.Trace.RunID != runID {
				return
			}
			switch completedMsg.Body.Event {
			case protocol.StreamEventRunCompleted, protocol.StreamEventRunFailed:
				select {
				case completedCh <- completedMsg:
				default:
				}
			}
		})
		if err != nil {
			t.Logf("completed topic subscription error: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-completedCh:
		if msg.Body.Event == protocol.StreamEventRunCompleted {
			payload, err := runCompletedPayloadFromCompletedMessage(msg)
			if err != nil {
				return err
			}
			if strings.TrimSpace(payload.Result.Message) == "" {
				return fmt.Errorf("completed payload message is empty")
			}
		}
		return nil
	}
}
func runCompletedPayloadFromCompletedMessage(msg protocol.MessageStreamMessage) (events.RunCompletedPayload, error) {
	if msg.Body.RunCompleted != nil {
		return *msg.Body.RunCompleted, nil
	}
	return events.RunCompletedPayload{}, fmt.Errorf("run completed payload is empty")
}

type workerHealthResponse struct {
	Status   string `json:"status"`
	Healthy  bool   `json:"healthy"`
	OrgID    uint   `json:"org_id"`
	WorkerID uint   `json:"worker_id"`
}

func checkWorkerHealthOrSkip(t *testing.T, healthURL string, wantOrgID uint, wantWorkerID uint) {
	t.Helper()

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		t.Skipf("skip real NATS publish test: worker health check unavailable at %s: %v", healthURL, err)
	}
	defer resp.Body.Close()

	var health workerHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Skipf("skip real NATS publish test: decode worker health response from %s: %v", healthURL, err)
	}
	if resp.StatusCode != http.StatusOK || !health.Healthy || strings.ToLower(health.Status) != "healthy" {
		t.Skipf("skip real NATS publish test: worker is not healthy: status_code=%d status=%q healthy=%t", resp.StatusCode, health.Status, health.Healthy)
	}
	if health.OrgID != wantOrgID || health.WorkerID != wantWorkerID {
		t.Skipf("skip real NATS publish test: worker identity mismatch: got org_id=%d worker_id=%d, want org_id=%d worker_id=%d",
			health.OrgID,
			health.WorkerID,
			wantOrgID,
			wantWorkerID,
		)
	}
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

func getenvDuration(key string, fallback time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		return fallback
	}
	return duration
}

func randomTestID(t *testing.T, prefix string) string {
	t.Helper()

	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		t.Fatalf("generate %s id: %v", prefix, err)
	}
	return fmt.Sprintf("%s_test_agent_run_%s", prefix, hex.EncodeToString(buf[:]))
}
