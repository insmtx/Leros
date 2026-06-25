package types

import (
	"gorm.io/gorm"
)

// ProjectFile 记录项目工作区的文件关联关系
type ProjectFile struct {
	gorm.Model
	PublicID        string         `gorm:"column:public_id;type:varchar(255);not null;uniqueIndex"`
	OrgID           uint           `gorm:"column:org_id;type:integer;not null;index"`
	ProjectID       uint           `gorm:"column:project_id;type:bigint;not null;index"`
	ProjectPublicID string         `gorm:"column:project_public_id;type:varchar(255);not null;index"`
	Filename        string         `gorm:"column:filename;type:varchar(500)"`
	OriginalName    string         `gorm:"column:original_name;type:varchar(500)"`
	MimeType        string         `gorm:"column:mime_type;type:varchar(100)"`
	FileSize        int64          `gorm:"column:file_size;type:bigint"`
	StoragePath     string         `gorm:"column:storage_path;type:varchar(500);not null"`
	Sha256          string         `gorm:"column:sha256;type:varchar(64);index"`
	Source          string         `gorm:"column:source;type:varchar(50);not null;default:'user_upload';index"`
	ArtifactID      *uint          `gorm:"column:artifact_id;type:bigint;index"`
	Metadata        ObjectMetadata `gorm:"column:metadata;type:jsonb"`
}

func (ProjectFile) TableName() string {
	return TableNameProjectFile
}
