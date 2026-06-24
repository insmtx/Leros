package types

import "gorm.io/gorm"

// MessageResource records which resource (skill/MCP/tool/etc.) a message used.
type MessageResource struct {
	gorm.Model
	ResourcePublicID string         `gorm:"column:resource_public_id;type:varchar(255);not null;uniqueIndex"`
	MessageID        uint           `gorm:"column:message_id;type:bigint;not null;index"`
	SessionID        uint           `gorm:"column:session_id;type:bigint;not null;index"`
	ResourceType     string         `gorm:"column:resource_type;type:varchar(50);not null;index"`
	ResourceCode     string         `gorm:"column:resource_code;type:varchar(255);not null;index"`
	ResourceName     string         `gorm:"column:resource_name;type:varchar(255);not null"`
	InvokeType       string         `gorm:"column:invoke_type;type:varchar(50);not null"`
	Seq              int            `gorm:"column:seq;type:integer;not null;default:0"`
	Meta             ObjectMetadata `gorm:"column:meta;type:jsonb"`
}

// TableName specifies the database table name for MessageResource.
func (MessageResource) TableName() string {
	return TableNameMessageResource
}
