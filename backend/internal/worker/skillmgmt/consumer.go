// Package skillmgmt provides a NATS consumer that handles unified skill management
// requests (install, list, uninstall) by shelling out to the leros CLI.
package skillmgmt

import (
	"bytes"
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

const consumerName = "worker-skill-mgmt"

// Config holds the configuration for a skill management consumer.
type Config struct {
	OrgID    uint
	WorkerID uint
}

// Consumer subscribes to skill management requests and dispatches by action.
type Consumer struct {
	cfg        Config
	subscriber eventbus.Subscriber
	conn       *nats.Conn
}

// New creates a new skill management consumer.
// The conn parameter is required for publishing reply messages via core NATS.
func New(cfg Config, subscriber eventbus.Subscriber, conn *nats.Conn) (*Consumer, error) {
	if cfg.OrgID == 0 {
		return nil, fmt.Errorf("org_id is required")
	}
	if cfg.WorkerID == 0 {
		return nil, fmt.Errorf("worker_id is required")
	}
	if subscriber == nil {
		return nil, fmt.Errorf("subscriber is required")
	}
	if conn == nil {
		return nil, fmt.Errorf("NATS connection is required")
	}
	return &Consumer{cfg: cfg, subscriber: subscriber, conn: conn}, nil
}

// Topic returns the NATS subject for this consumer.
func (c *Consumer) Topic() string {
	topic, err := dm.WorkerSkillSubject(c.cfg.OrgID, c.cfg.WorkerID)
	if err != nil {
		logs.Errorf("Failed to build skill management topic: %v", err)
		return ""
	}
	return topic
}

// Start subscribes to the skill management topic and processes incoming requests.
func (c *Consumer) Start(ctx context.Context) error {
	topic := c.Topic()
	logs.InfoContextf(ctx, "Starting skill management subscription: %s", topic)
	return c.subscriber.Subscribe(ctx, topic, consumerName, func(msg *nats.Msg) {
		if err := c.handle(ctx, msg); err != nil {
			logs.ErrorContextf(ctx, "Failed to handle skill management request: %v", err)
		}
	})
}

func (c *Consumer) handle(ctx context.Context, msg *nats.Msg) error {
	var req protocol.SkillManagementMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return c.replyError("", "unmarshal request", err)
	}

	action := strings.TrimSpace(req.Body.Action)
	logs.InfoContextf(ctx,
		"Received skill management request: action=%s msg_id=%s org_id=%d worker_id=%d reply_to=%s",
		action, req.ID, req.Route.OrgID, req.Route.WorkerID, req.Body.ReplyTo,
	)

	switch action {
	case "install":
		return c.handleInstall(ctx, req)
	case "list":
		return c.handleList(ctx, req)
	case "uninstall":
		return c.handleUninstall(ctx, req)
	default:
		return c.replyError(req.Body.ReplyTo, fmt.Sprintf("unknown action: %s", action), nil)
	}
}

func (c *Consumer) handleInstall(ctx context.Context, req protocol.SkillManagementMessage) error {
	skillID := strings.TrimSpace(req.Body.SkillID)
	if skillID == "" {
		return c.replyError(req.Body.ReplyTo, "skill_id is empty", nil)
	}

	lerosBin, err := os.Executable()
	if err != nil {
		return c.replyError(req.Body.ReplyTo, "find leros binary", err)
	}

	cmd := exec.CommandContext(ctx, lerosBin, "skill", "install", skillID, "--force", "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logs.InfoContextf(ctx, "Running: %s skill install %s --force --yes", lerosBin, skillID)
	if err := cmd.Run(); err != nil {
		logs.ErrorContextf(ctx, "leros skill install failed for %q: %v", skillID, err)
		return c.replyError(req.Body.ReplyTo, fmt.Sprintf("install skill %q", skillID), err)
	}

	logs.InfoContextf(ctx, "leros skill install succeeded for %q", skillID)
	return c.replySuccess(req.Body.ReplyTo, "install", fmt.Sprintf("skill %q installed", skillID))
}

func (c *Consumer) handleList(ctx context.Context, req protocol.SkillManagementMessage) error {
	lerosBin, err := os.Executable()
	if err != nil {
		return c.replyError(req.Body.ReplyTo, "find leros binary", err)
	}

	cmd := exec.CommandContext(ctx, lerosBin, "skill", "list", "--json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		errDetail := stderr.String()
		if errDetail == "" {
			if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
				errDetail = string(exitErr.Stderr)
			}
		}
		if errDetail != "" {
			errDetail = strings.TrimSpace(errDetail)
			logs.ErrorContextf(ctx, "leros skill list failed: %v, stderr: %s", err, errDetail)
			return c.replyError(req.Body.ReplyTo, fmt.Sprintf("list skills: %s", errDetail), err)
		}
		logs.ErrorContextf(ctx, "leros skill list failed: %v", err)
		return c.replyError(req.Body.ReplyTo, "list skills", err)
	}

	var items []protocol.SkillListItem
	if err := json.Unmarshal(output, &items); err != nil {
		logs.ErrorContextf(ctx, "Failed to unmarshal skill list output: %v, raw=%s", err, string(output))
		return c.replyError(req.Body.ReplyTo, "unmarshal list output", err)
	}

	resp := protocol.SkillManagementResponse{
		Success: true,
		Action:  "list",
		Data:    items,
	}
	return c.publishReply(req.Body.ReplyTo, resp)
}

func (c *Consumer) handleUninstall(ctx context.Context, req protocol.SkillManagementMessage) error {
	name := strings.TrimSpace(req.Body.Name)
	if name == "" {
		return c.replyError(req.Body.ReplyTo, "name is empty", nil)
	}

	lerosBin, err := os.Executable()
	if err != nil {
		return c.replyError(req.Body.ReplyTo, "find leros binary", err)
	}

	cmd := exec.CommandContext(ctx, lerosBin, "skill", "uninstall", name, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logs.InfoContextf(ctx, "Running: %s skill uninstall %s --yes", lerosBin, name)
	if err := cmd.Run(); err != nil {
		logs.ErrorContextf(ctx, "leros skill uninstall failed for %q: %v", name, err)
		return c.replyError(req.Body.ReplyTo, fmt.Sprintf("uninstall skill %q", name), err)
	}

	logs.InfoContextf(ctx, "leros skill uninstall succeeded for %q", name)
	return c.replySuccess(req.Body.ReplyTo, "uninstall", fmt.Sprintf("skill %q uninstalled", name))
}

// replySuccess publishes a success response to the reply inbox.
func (c *Consumer) replySuccess(replyTo, action, message string) error {
	resp := protocol.SkillManagementResponse{
		Success: true,
		Action:  action,
		Message: message,
	}
	return c.publishReply(replyTo, resp)
}

// replyError publishes an error response to the reply inbox.
func (c *Consumer) replyError(replyTo, context string, err error) error {
	errMsg := context
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", context, err)
	}
	resp := protocol.SkillManagementResponse{
		Success: false,
		Error:   errMsg,
	}
	if pubErr := c.publishReply(replyTo, resp); pubErr != nil {
		return fmt.Errorf("%s (and failed to publish reply: %v)", errMsg, pubErr)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", context, err)
	}
	return fmt.Errorf("%s", context)
}

// publishReply publishes a response to the given reply inbox via core NATS.
func (c *Consumer) publishReply(replyTo string, resp protocol.SkillManagementResponse) error {
	if replyTo == "" {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal reply: %w", err)
	}
	return c.conn.Publish(replyTo, data)
}
