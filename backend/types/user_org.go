package types

import "gorm.io/gorm"

// TableNameUserOrg 用户组织关联表名
const TableNameUserOrg = "user_orgs"

// UserOrg 表示用户与组织的关联关系
// 该表是多对多关系的中间表，每个关联记录有唯一的 uin
type UserOrg struct {
	gorm.Model
	Uin       uint `gorm:"column:uin;type:bigint;unique_index;not null"` // 关联ID，JWT中的Uin
	UserID    uint `gorm:"column:user_id;type:bigint;index;not null"`    // 用户ID
	OrgID     uint `gorm:"column:org_id;type:bigint;index;not null"`     // 组织ID
	IsDefault bool `gorm:"column:is_default;type:boolean;default:false"` // 是否为默认组织
}

// TableName 重写表名
func (UserOrg) TableName() string {
	return TableNameUserOrg
}
