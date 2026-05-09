package types

import "gorm.io/gorm"

// TableNameOrganization 组织表名
const TableNameOrganization = "organizations"

// Organization 表示系统中的组织/企业信息
type Organization struct {
	gorm.Model
	Code string `gorm:"column:code;type:varchar(255);unique_index;not null"` // 组织代码（唯一）
	Name string `gorm:"column:name;type:varchar(255);not null"`               // 组织名称
}

// TableName 重写表名
func (Organization) TableName() string {
	return TableNameOrganization
}
