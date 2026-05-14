// types 包提供 Leros 的核心数据类型定义
//
// 该包定义了数字助手、事件、用户、技能等核心领域模型，
// 以及相关的常量和数据库表名定义。
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// LLMModelStatus 表示LLM模型配置的启用状态
type LLMModelStatus string

const (
	// LLMModelStatusActive 表示模型配置可被业务使用
	LLMModelStatusActive LLMModelStatus = "active"
	// LLMModelStatusInactive 表示模型配置暂不可被业务使用
	LLMModelStatusInactive LLMModelStatus = "inactive"
)

// LLMModel 定义一个在数据库中持久化的LLM模型配置
//
// 该表以“可直接调用的模型配置”为最小管理单元，暂时将供应商、
// 模型、凭证、默认参数和扩展配置放在同一张表中，便于MVP阶段快速接入多模型配置。
type LLMModel struct {
	gorm.Model

	// 模型所属组织ID，用于隔离不同组织的模型配置
	OrgID uint `gorm:"column:org_id;type:integer;not null;index;uniqueIndex:idx_llm_model_org_code;uniqueIndex:idx_llm_model_org_default,where:is_default = true AND deleted_at IS NULL"`
	// 模型配置编码，在组织内唯一，用于业务配置引用
	Code string `gorm:"column:code;type:varchar(128);not null;uniqueIndex:idx_llm_model_org_code"`
	// 模型展示名称，用于前端列表和选择器展示
	Name string `gorm:"column:name;type:varchar(255);not null"`
	// 模型描述，用于记录用途、额度、注意事项等说明
	Description string `gorm:"column:description;type:text"`

	// 模型供应商标识，例如 openai、anthropic、deepseek、custom
	Provider string `gorm:"column:provider;type:varchar(64);not null;index"` // 建议使用 types.LLMProviderType 定义的常量值
	// 实际传递给供应商API或CLI的模型名称
	ModelName string `gorm:"column:model;type:varchar(255);not null"`
	// API基础地址，官方默认地址可为空，自定义网关或兼容接口需填写
	BaseURL string `gorm:"column:base_url;type:varchar(500)"`

	// API Key密文，业务读取时需要通过统一密钥服务解密
	APIKeyEncrypted string `gorm:"column:api_key_encrypted;type:text"`
	// API Key脱敏展示值，仅用于前端展示，禁止用于真实调用
	APIKeyMasked string `gorm:"column:api_key_masked;type:varchar(128)"`

	// 默认最大输出Token数，运行请求可按需覆盖
	MaxTokens int `gorm:"column:max_tokens;type:integer;default:4096"`
	// 默认采样温度，运行请求可按需覆盖
	Temperature float64 `gorm:"column:temperature;type:decimal(4,3);default:0.7"`
	// 默认请求超时时间，单位秒
	TimeoutSec int `gorm:"column:timeout_sec;type:integer;default:120"`

	// 模型状态，inactive状态下不应被新任务选择
	Status string `gorm:"column:status;type:varchar(32);not null;default:active;index"` // 建议使用 types.LLMModelStatus 定义的常量值
	// 是否为组织默认模型，业务未显式选择模型时可回退到该模型
	IsDefault bool `gorm:"column:is_default;type:boolean;default:false;index"`
	// 是否为系统内置配置，内置配置可用于初始化或演示场景
	IsSystem bool `gorm:"column:is_system;type:boolean;default:false"`

	// 扩展配置，保存能力标记、上下文窗口、价格、额外Header等非核心字段
	Config LLMModelConfig `gorm:"column:config;type:jsonb"`
}

// TableName 指定LLMModel对应的数据库表名
func (LLMModel) TableName() string {
	return TableNameLLMModel
}

// LLMModelConfig 表示LLM模型配置的扩展JSON字段
type LLMModelConfig map[string]interface{}

// Scan 实现 sql.Scanner 接口，用于从数据库中读取 JSON 数据
func (c *LLMModelConfig) Scan(value interface{}) error {
	if value == nil {
		*c = LLMModelConfig{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into LLMModelConfig", value)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*c = LLMModelConfig(result)
	return nil
}

// Value 实现 driver.Valuer 接口，用于将扩展配置转换为 JSON 存储
func (c LLMModelConfig) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(map[string]interface{}(c))
}
