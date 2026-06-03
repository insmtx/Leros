package runtime

import (
	"context"

	"github.com/insmtx/Leros/backend/engines"
	"github.com/ygpkg/yg-go/logs"
)

// autoApprovalHandler 模拟审批处理器：打印审批信息并自动批准。
type autoApprovalHandler struct{}

func (h *autoApprovalHandler) RequestApproval(ctx context.Context, req *engines.ApprovalRequest) (*engines.ApprovalDecision, error) {
	logs.InfoContextf(ctx, "[APPROVAL-MOCK] request_id=%s tool=%s description=%s engine=%s → auto approved",
		req.RequestID, req.ToolName, req.Description, req.Engine)
	return &engines.ApprovalDecision{
		RequestID: req.RequestID,
		Action:    "approved",
		Reason:    "auto-approved by mock handler",
	}, nil
}

var _ engines.ApprovalHandler = (*autoApprovalHandler)(nil)
