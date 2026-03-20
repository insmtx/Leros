package types

import "gorm.io/gorm"

// SkillExecutionLog 记录技能执行日志
type SkillExecutionLog struct {
	gorm.Model
	// 关联的技能ID
	SkillID uint `gorm:"column:skill_id;type:integer;not null;index"`
	// 执行会话ID
	SessionID string `gorm:"column:session_id;type:varchar(255);index"`
	// 用户ID
	UserID string `gorm:"column:user_id;type:varchar(255);index"`
	// 数字助手ID
	AssistantID string `gorm:"column:assistant_id;type:varchar(255);index"`
	// 输入参数
	Input map[string]interface{} `gorm:"column:input;type:jsonb"`
	// 输出结果
	Output map[string]interface{} `gorm:"column:output;type:jsonb"`
	// 是否成功
	Success bool `gorm:"column:success;type:boolean"`
	// 错误信息
	ErrorMsg string `gorm:"column:error_msg;type:text"`
	// 执行耗时（毫秒）
	Duration int64 `gorm:"column:duration;type:bigint"`
}

// TableName 指定SkillExecutionLog对应的数据库表名
func (SkillExecutionLog) TableName() string {
	return TableNameSkillLog
}
