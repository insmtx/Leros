package service

import (
	"context"
)

type AssistantInferrer interface {
	InferAssignedAssistantID(ctx context.Context, sessionOrgID uint, sessionType string) uint
}

type DefaultAssistantInferrer struct {
	defaultAssistantID uint
}

func NewDefaultAssistantInferrer(defaultAssistantID uint) *DefaultAssistantInferrer {
	return &DefaultAssistantInferrer{
		defaultAssistantID: defaultAssistantID,
	}
}

func (d *DefaultAssistantInferrer) InferAssignedAssistantID(ctx context.Context, sessionOrgID uint, sessionType string) uint {
	// TODO: 实现基于规则的智能推断逻辑
	// 考虑因素：
	// 1. sessionOrgID - 组织 ID 可用于选择组织级别的默认数字员工
	// 2. sessionType - 会话类型（user_chat, assistant_instance）可用于路由到不同类型的助手
	// 3. 用户历史偏好 - 可根据 uin 查询历史分配记录
	// 4. 负载平衡 - 在多数字员工场景下平衡分配

	if d.defaultAssistantID == 0 {
		return 1
	}
	return d.defaultAssistantID
}
