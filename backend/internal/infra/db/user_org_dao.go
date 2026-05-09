package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/types"
)

func GetUserOrgByUin(ctx context.Context, db *gorm.DB, uin uint) (*types.UserOrg, error) {
	var userOrg types.UserOrg
	result := db.WithContext(ctx).Where("uin = ?", uin).First(&userOrg)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userOrg, nil
}

func GetUserOrgByUserID(ctx context.Context, db *gorm.DB, userID uint) (*types.UserOrg, error) {
	var userOrg types.UserOrg
	result := db.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&userOrg)
	if result.Error != nil {
		result = db.WithContext(ctx).Where("user_id = ?", userID).First(&userOrg)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return &userOrg, nil
}

func CreateUserOrg(ctx context.Context, db *gorm.DB, userOrg *types.UserOrg) error {
	return db.WithContext(ctx).Create(userOrg).Error
}

func DeleteUserOrg(ctx context.Context, db *gorm.DB, id uint) error {
	return db.WithContext(ctx).Delete(&types.UserOrg{}, id).Error
}
