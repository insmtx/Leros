package types

import "gorm.io/gorm"

type FileUpload struct {
	gorm.Model
	PublicID     string         `gorm:"column:public_id;type:varchar(255);not null;uniqueIndex"`
	OrgID        uint           `gorm:"column:org_id;type:integer;not null;index"`
	OwnerID      uint           `gorm:"column:owner_id;type:bigint;not null;index"`
	Filename     string         `gorm:"column:filename;type:varchar(500);not null"`
	OriginalName string         `gorm:"column:original_name;type:varchar(500);not null"`
	MimeType     string         `gorm:"column:mime_type;type:varchar(100)"`
	FileSize     int64          `gorm:"column:file_size;type:bigint"`
	StoragePath  string         `gorm:"column:storage_path;type:varchar(2000);not null"`
	Sha256       string         `gorm:"column:sha256;type:varchar(64)"`
	Purpose      string         `gorm:"column:purpose;type:varchar(50);index"`
	Status       string         `gorm:"column:status;type:varchar(50);default:'active'"`
	Metadata     ObjectMetadata `gorm:"column:metadata;type:jsonb"`
}

func (FileUpload) TableName() string {
	return TableNameFileUpload
}
