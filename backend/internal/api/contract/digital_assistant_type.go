package contract

import "time"

// DigitalAssistantStatus 数字助手状态常量
type DigitalAssistantStatus string

const (
	DigitalAssistantStatusDraft     DigitalAssistantStatus = "draft"
	DigitalAssistantStatusActive    DigitalAssistantStatus = "active"
	DigitalAssistantStatusInactive  DigitalAssistantStatus = "inactive"
	DigitalAssistantStatusArchived  DigitalAssistantStatus = "archived"
)

// RuntimeType 运行时类型常量
type RuntimeType string

const (
	RuntimeTypeDocker   RuntimeType = "docker"
	RuntimeTypeProcess  RuntimeType = "process"
	RuntimeTypeK8s      RuntimeType = "kubernetes"
)

// LLMProviderType LLM提供商类型常量
type LLMProviderType string

const (
	LLMProviderOpenAI    LLMProviderType = "openai"
	LLMProviderClaude    LLMProviderType = "claude"
	LLMProviderDeepSeek  LLMProviderType = "deepseek"
)

// MemoryType 记忆类型常量
type MemoryType string

const (
	MemoryTypeRedis    MemoryType = "redis"
	MemoryTypePostgres MemoryType = "postgres"
)

// ChannelType 渠道类型常量
type ChannelType string

const (
	ChannelTypeGitHub  ChannelType = "github"
	ChannelTypeGitLab  ChannelType = "gitlab"
	ChannelTypeWeChat  ChannelType = "wechat"
	ChannelTypeFeishu  ChannelType = "feishu"
)

// KnowledgeType 知识库类型常量
type KnowledgeType string

const (
	KnowledgeTypeVector  KnowledgeType = "vector"
	KnowledgeTypeFile    KnowledgeType = "file"
	KnowledgeTypeDatabase KnowledgeType = "database"
)

// DigitalAssistant 数字助手信息
type DigitalAssistant struct {
	ID          uint               `json:"id"`
	Code        string             `json:"code"`
	OrgID       uint               `json:"org_id"`
	OwnerID     uint               `json:"owner_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Avatar      string             `json:"avatar"`
	Status      string             `json:"status"`
	Version     int                `json:"version"`
	Config      AssistantConfig    `json:"config"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// AssistantConfig 数字助手配置
type AssistantConfig struct {
	Runtime  RuntimeConfig  `json:"runtime_config"`
	LLM      LLMConfig      `json:"llm_config"`
	Skills   []SkillRef     `json:"skills"`
	Channels []ChannelRef   `json:"channels"`
	Knowledge []KnowledgeRef `json:"knowledge"`
	Memory   MemoryConfig   `json:"memory_config"`
	Policies PolicyConfig   `json:"policies_config"`
}

// SkillRef 技能引用
type SkillRef struct {
	SkillCode string `json:"skill_code"`
	Version   string `json:"version"`
}

// ChannelRef 渠道引用
type ChannelRef struct {
	Type string `json:"type"`
}

// KnowledgeRef 知识库引用
type KnowledgeRef struct {
	Type     string `json:"type"`
	DatasetID string `json:"dataset_id"`
	Repo     string `json:"repo"`
}

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	Type string `json:"type"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Type string `json:"type"`
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	Type string `json:"type"`
}

// PolicyConfig 策略配置
type PolicyConfig struct {
	Type string `json:"type"`
}

// CreateDigitalAssistantRequest 创建数字助手请求
type CreateDigitalAssistantRequest struct {
	Code        string          `json:"code" binding:"required"`
	OrgID       uint            `json:"org_id" binding:"required"`
	OwnerID     uint            `json:"owner_id" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Avatar      string          `json:"avatar"`
	Config      AssistantConfig `json:"config" binding:"required"`
}

// UpdateDigitalAssistantRequest 更新数字助手请求
type UpdateDigitalAssistantRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Avatar      string          `json:"avatar"`
	Config      *AssistantConfig `json:"config,omitempty"`
}

// UpdateDigitalAssistantStatusRequest 更新数字助手状态请求
type UpdateDigitalAssistantStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ListDigitalAssistantRequest 查询数字助手列表请求
type ListDigitalAssistantRequest struct {
	OrgID   *uint    `form:"org_id,omitempty"`
	OwnerID *uint    `form:"owner_id,omitempty"`
	Status  *string  `form:"status,omitempty"`
	Keyword *string  `form:"keyword,omitempty"`
	Page    int      `form:"page,default=1"`
	PerPage int      `form:"per_page,default=20"`
}

// DigitalAssistantList 数字助手列表响应
type DigitalAssistantList struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Items []DigitalAssistant `json:"items"`
}

// DigitalAssistantDetail 数字助手详情响应
type DigitalAssistantDetail struct {
	DigitalAssistant
}

// UpdateDigitalAssistantConfigRequest 更新数字助手配置请求
type UpdateDigitalAssistantConfigRequest struct {
	Config AssistantConfig `json:"config" binding:"required"`
}
