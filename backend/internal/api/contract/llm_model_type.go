package contract

import "time"

// LLMModel LLM模型配置响应结构
type LLMModel struct {
	ID          uint                   `json:"id"`
	OrgID       uint                   `json:"org_id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	BaseURL     string                 `json:"base_url"`
	APIKey      string                 `json:"api_key"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature"`
	TimeoutSec  int                    `json:"timeout_sec"`
	Status      string                 `json:"status"`
	IsDefault   bool                   `json:"is_default"`
	IsSystem    bool                   `json:"is_system"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateLLMModelRequest 创建LLM模型配置请求
type CreateLLMModelRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Provider    string                 `json:"provider,omitempty"`
	Model       string                 `json:"model" binding:"required"`
	BaseURL     string                 `json:"base_url" binding:"required"`
	APIKey      string                 `json:"api_key" binding:"required"`
	Status      string                 `json:"status,omitempty"`
	IsDefault   bool                   `json:"is_default,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// UpdateLLMModelRequest 更新LLM模型配置请求
type UpdateLLMModelRequest struct {
	Name        string                  `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Provider    string                  `json:"provider,omitempty"`
	Model       string                  `json:"model,omitempty"`
	BaseURL     *string                 `json:"base_url,omitempty"`
	APIKey      *string                 `json:"api_key,omitempty"`
	Status      string                  `json:"status,omitempty"`
	IsDefault   *bool                   `json:"is_default,omitempty"`
	Config      *map[string]interface{} `json:"config,omitempty"`
}

// ListLLMModelsRequest 查询LLM模型配置列表请求
type ListLLMModelsRequest struct {
	Provider *string `json:"provider,omitempty" form:"provider,omitempty"`
	Status   *string `json:"status,omitempty" form:"status,omitempty"`
	Keyword  *string `json:"keyword,omitempty" form:"keyword,omitempty"`
	Page     int     `json:"page,omitempty" form:"page,default=1"`
	PerPage  int     `json:"per_page,omitempty" form:"per_page,default=20"`
}

// LLMModelList LLM模型配置列表响应
type LLMModelList struct {
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Items []LLMModel `json:"items"`
}

// TestLLMModelRequest 测试LLM模型配置请求
type TestLLMModelRequest struct {
	ID       *uint  `json:"id,omitempty"`
	Code     string `json:"code,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// TestLLMModelResponse 测试LLM模型配置响应
type TestLLMModelResponse struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Endpoint   string `json:"endpoint"`
	LatencyMS  int64  `json:"latency_ms"`
}
