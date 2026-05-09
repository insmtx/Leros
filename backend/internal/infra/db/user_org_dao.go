package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/types"
)

// GetUserOrgByUin 根据UIN获取用户组织
func GetUserOrgByUin(ctx context.Context, db *gorm.DB, uin uint) (*types.UserOrg, error) {
	var userOrg types.UserOrg
	err := db.WithContext(ctx).Where("uin = ?", uin).First(&userOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userOrg, nil
}

// GetUserOrgByUserID 获取用户默认组织（若无默认则取首个）
func GetUserOrgByUserID(ctx context.Context, db *gorm.DB, userID uint) (*types.UserOrg, error) {
	var userOrg types.UserOrg
	// 优先获取默认组织
	err := db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&userOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 若无默认组织，获取首个组织
			err = db.WithContext(ctx).Where("user_id = ?", userID).First(&userOrg).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, nil
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &userOrg, nil
}

// CreateUserOrg 创建用户组织
func CreateUserOrg(ctx context.Context, db *gorm.DB, userOrg *types.UserOrg) error {
	return db.WithContext(ctx).Create(userOrg).Error
}

// DeleteUserOrg 删除用户组织
func DeleteUserOrg(ctx context.Context, db *gorm.DB, id uint) error {
	return db.WithContext(ctx).Delete(&types.UserOrg{}, id).Error
}
