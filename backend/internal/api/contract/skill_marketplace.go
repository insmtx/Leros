package contract

import (
	"context"
	"io"
)

// SkillPackageDownload 技能包下载结果。
type SkillPackageDownload struct {
	Reader   io.ReadCloser
	FileName string
}

// InstallSkillRequest Skill 安装请求。
type InstallSkillRequest struct {
	Source  string `json:"source" binding:"required"`
	SkillID string `json:"skill_id" binding:"required"`
}

// InstallSkillResponse Skill 安装响应（异步，仅表示已接受）。
type InstallSkillResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SkillMarketplaceService 定义 Skill 市场搜索服务接口。
type SkillMarketplaceService interface {
	SearchSkillMarketplace(ctx context.Context, req *SearchSkillMarketplaceRequest) (*SearchSkillMarketplaceResponse, error)
	DownloadBuiltinSkill(ctx context.Context, skillID string) (*SkillPackageDownload, error)
	InstallSkill(ctx context.Context, req *InstallSkillRequest) (*InstallSkillResponse, error)
}

// SearchSkillMarketplaceRequest Skill 市场搜索请求。
type SearchSkillMarketplaceRequest struct {
	Keyword     string   `form:"keyword" json:"keyword,omitempty"`
	Category    string   `form:"category" json:"category,omitempty"`
	SourceTypes []string `form:"source_types" json:"source_types,omitempty"`
	Limit       int      `form:"limit" json:"limit,omitempty"`
}

// SkillMarketplaceItemView Skill 市场条目视图。
type SkillMarketplaceItemView struct {
	SourceType  string   `json:"source_type"`
	SkillID     string   `json:"skill_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Icon        string   `json:"icon,omitempty"`
	Installs    int64    `json:"installs"`
}

// SkillSourceWarning 源查询警告信息。
type SkillSourceWarning struct {
	SourceType string `json:"source_type"`
	Message    string `json:"message"`
}

// SearchSkillMarketplaceResponse Skill 市场搜索响应。
type SearchSkillMarketplaceResponse struct {
	Items    []SkillMarketplaceItemView `json:"items"`
	Warnings []SkillSourceWarning       `json:"warnings,omitempty"`
}
