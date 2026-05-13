// db 包提供 Leros 的数据库初始化和管理功能
//
// 该包负责数据库连接的初始化、表结构的自动迁移，
// 以及提供获取数据库实例的方法。
package db

import (
	"fmt"

	"github.com/ygpkg/yg-go/dbtools"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/types"
)

// dbName 是数据库名称常量
const dbName = "singer"

// InitDB 创建并初始化数据库连接
//
// 使用 dbtools 初始化数据库连接，并根据配置决定是否启用调试模式，
// 最后运行数据库迁移来创建所有必要的表结构。
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := dbtools.InitDBConn(dbName, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if cfg.Debug {
		db = db.Debug()
	}

	// 运行数据库迁移
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// 初始化开发数据（默认组织、用户、用户组织关联）
	if err := InitDevData(db); err != nil {
		return nil, fmt.Errorf("failed to init dev data: %w", err)
	}

	logs.Info("Database connection initialized successfully")
	return db, nil
}

// runMigrations 为所有模型创建数据库表
//
// 该函数会自动为所有定义的模型创建或更新数据库表结构。
func runMigrations(db *gorm.DB) error {
	models := []interface{}{
		&types.User{},
		&types.Event{},
		&types.DigitalAssistant{},
		&types.Skill{},
		&types.SkillRegistry{},
		&types.SkillExecutionLog{},
		&types.Session{},
		&types.SessionMessage{},
		&types.Organization{},
		&types.UserOrg{},
		&types.LLMModel{},
	}

	if err := dbtools.InitModel(db, models...); err != nil {
		return err
	}

	logs.Info("Database migrations completed")
	return nil
}

// InitDevData 初始化开发环境数据（仅在数据为空时执行）
// 包括：默认组织、默认用户、用户组织关联
func InitDevData(db *gorm.DB) error {
	// 初始化默认组织
	var orgCount int64
	db.Model(&types.Organization{}).Count(&orgCount)
	if orgCount == 0 {
		defaultOrg := &types.Organization{
			Code:   "default_org",
			Name:   "默认组织",
			Type:   "company",
			Status: "active",
		}
		if err := db.Create(defaultOrg).Error; err != nil {
			return fmt.Errorf("failed to create default org: %w", err)
		}
		logs.Info("Default organization created")
	}

	// 初始化默认用户
	var userCount int64
	db.Model(&types.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin123456"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		defaultUser := &types.User{
			GithubID:    0,
			GithubLogin: "admin",
			Name:        "Admin User",
			Email:       "admin@singer.local",
			Password:    string(hashedPassword),
		}
		if err := db.Create(defaultUser).Error; err != nil {
			return fmt.Errorf("failed to create default user: %w", err)
		}
		logs.Info("Default user created (login: admin)")
	}

	// 初始化用户组织关联
	var userOrgCount int64
	db.Model(&types.UserOrg{}).Count(&userOrgCount)
	if userOrgCount == 0 {
		var user types.User
		var org types.Organization
		if err := db.Where("github_login = ?", "admin").First(&user).Error; err != nil {
			return fmt.Errorf("failed to find default user: %w", err)
		}
		if err := db.Where("code = ?", "default_org").First(&org).Error; err != nil {
			return fmt.Errorf("failed to find default org: %w", err)
		}

		userOrg := &types.UserOrg{
			Uin:       user.ID,
			UserID:    user.ID,
			OrgID:     org.ID,
			IsDefault: true,
		}
		if err := db.Create(userOrg).Error; err != nil {
			return fmt.Errorf("failed to create default user-org: %w", err)
		}
		logs.Infof("Default user-org association created (uin=%d, user_id=%d, org_id=%d)", userOrg.Uin, userOrg.UserID, userOrg.OrgID)
	}

	return nil
}

// GetDB 获取默认的数据库实例
func GetDB() *gorm.DB {
	return dbtools.DB(dbName)
}
