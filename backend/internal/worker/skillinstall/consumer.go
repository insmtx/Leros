// Package skillinstall provides a lightweight NATS consumer that handles skill
// installation requests by shelling out to the leros CLI.
package skillinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nats-io/nats.go"

	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/ygpkg/yg-go/logs"
)

const consumerName = "worker-skill-install"

// Config holds the configuration for a skill install consumer.
type Config struct {
	OrgID    uint
	WorkerID uint
}

// Consumer subscribes to skill install requests and runs leros skill install.
type Consumer struct {
	cfg        Config
	subscriber eventbus.Subscriber
}

// New creates a new skill install consumer.
func New(cfg Config, subscriber eventbus.Subscriber) (*Consumer, error) {
	if cfg.OrgID == 0 {
		return nil, fmt.Errorf("org_id is required")
	}
	if cfg.WorkerID == 0 {
		return nil, fmt.Errorf("worker_id is required")
	}
	if subscriber == nil {
		return nil, fmt.Errorf("subscriber is required")
	}
	return &Consumer{cfg: cfg, subscriber: subscriber}, nil
}

// Topic returns the NATS subject for this consumer.
func (c *Consumer) Topic() string {
	topic, err := dm.WorkerSkillInstallSubject(c.cfg.OrgID, c.cfg.WorkerID)
	if err != nil {
		logs.Errorf("Failed to build skill install topic: %v", err)
		return ""
	}
	return topic
}

// Start subscribes to the skill install topic and processes incoming requests.
func (c *Consumer) Start(ctx context.Context) error {
	topic := c.Topic()
	logs.InfoContextf(ctx, "Starting skill install subscription: %s", topic)
	return c.subscriber.Subscribe(ctx, topic, consumerName, func(msg *nats.Msg) {
		if err := c.handle(ctx, msg); err != nil {
			logs.ErrorContextf(ctx, "Failed to handle skill install: %v", err)
		}
	})
}

func (c *Consumer) handle(ctx context.Context, msg *nats.Msg) error {
	var taskMsg protocol.SkillInstallMessage
	if err := json.Unmarshal(msg.Data, &taskMsg); err != nil {
		return fmt.Errorf("unmarshal skill install message: %w", err)
	}

	body := taskMsg.Body
	skillID := strings.TrimSpace(body.SkillID)
	if skillID == "" {
		return fmt.Errorf("skill_id is empty")
	}

	logs.InfoContextf(ctx,
		"Received skill install request: source=%s skill_id=%s msg_id=%s org_id=%d worker_id=%d",
		body.Source, skillID, taskMsg.ID, taskMsg.Route.OrgID, taskMsg.Route.WorkerID,
	)

	lerosBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find leros binary: %w", err)
	}

	cmd := exec.CommandContext(ctx, lerosBin, "skill", "install", skillID, "--force", "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logs.InfoContextf(ctx, "Running: %s skill install %s --force --yes", lerosBin, skillID)
	if err := cmd.Run(); err != nil {
		logs.ErrorContextf(ctx, "leros skill install failed for %q: %v", skillID, err)
		return fmt.Errorf("leros skill install %q: %w", skillID, err)
	}

	logs.InfoContextf(ctx, "leros skill install succeeded for %q", skillID)
	return nil
}
