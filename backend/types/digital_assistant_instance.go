package types

import "gorm.io/gorm"

// TableName 指定DigitalAssistantInstance结构体对应的数据库表名
func (DigitalAssistantInstance) TableName() string {
	return TableNameDigitalAssistantInstance
}

type DigitalAssistantInstance struct {
	gorm.Model
}
