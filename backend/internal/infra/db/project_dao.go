package db

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/Leros/backend/types"
)

// CreateProject 创建项目
func CreateProject(ctx context.Context, db *gorm.DB, project *types.Project) error {
	return db.WithContext(ctx).Create(project).Error
}

// GetProjectByPublicID 根据组织ID和PublicID获取项目
func GetProjectByPublicID(ctx context.Context, db *gorm.DB, orgID uint, publicID string) (*types.Project, error) {
	var entity types.Project
	err := db.WithContext(ctx).Where("org_id = ? AND public_id = ?", orgID, publicID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// UpdateProject 更新项目
func UpdateProject(ctx context.Context, db *gorm.DB, project *types.Project) error {
	return db.WithContext(ctx).Save(project).Error
}

// DeleteProject 删除项目（软删除）
func DeleteProject(ctx context.Context, db *gorm.DB, id uint) error {
	return db.WithContext(ctx).Delete(&types.Project{}, id).Error
}

// CreateProjectMember 创建项目成员
func CreateProjectMember(ctx context.Context, db *gorm.DB, member *types.ProjectMember) error {
	return db.WithContext(ctx).Create(member).Error
}

// GetProjectSession 根据项目ID获取scope=project的会话
func GetProjectSession(ctx context.Context, db *gorm.DB, projectID uint) (*types.Session, error) {
	var entity types.Session
	err := db.WithContext(ctx).
		Where("project_id = ? AND type = ?", projectID, string(types.SessionTypeProject)).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// ListProjects 查询项目列表，使用 PageQuery 作为查询参数
func ListProjects(ctx context.Context, d *gorm.DB, opt *types.PageQuery) ([]*types.Project, int64, error) {
	var entities []*types.Project
	var total int64

	query := d.WithContext(ctx).Table(types.TableNameProject).
		Where("org_id = ? AND deleted_at IS NULL", opt.OrgID)

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			if filter.ExactMatch {
				query = query.Where("name IN (?)", filter.Value)
			} else {
				query = query.Where("name LIKE ?", "%"+filter.Value[0]+"%")
			}
		case "status":
			query = query.Where("status IN (?)", filter.Value)
		case "public_id":
			query = query.Where("public_id IN (?)", filter.Value)
		default:
			logs.WarnContextf(ctx, "[project][ListProjects] invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	} else {
		query = query.Order("created_at DESC")
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	} else {
		query = query.Limit(150)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	return entities, total, nil
}
