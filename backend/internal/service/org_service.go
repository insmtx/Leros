package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	"github.com/insmtx/Leros/backend/types"
)

var _ contract.OrgService = (*orgService)(nil)

type orgService struct {
	db *gorm.DB
}

func NewOrgService(d *gorm.DB) contract.OrgService {
	return &orgService{db: d}
}

func (s *orgService) CreateOrg(ctx context.Context, req *contract.CreateOrgRequest) (*contract.Org, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}

	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return nil, errors.New("code is required")
	}

	existing, err := db.GetOrgByCode(ctx, s.db, strings.TrimSpace(req.Code))
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("org code already exists")
	}

	orgType := strings.TrimSpace(req.Type)
	if orgType == "" {
		orgType = "company"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	org := &types.Organization{
		Type:   orgType,
		Code:   strings.TrimSpace(req.Code),
		Name:   strings.TrimSpace(req.Name),
		Status: status,
	}

	if err := db.CreateOrg(ctx, s.db, org); err != nil {
		return nil, err
	}

	return convertToContractOrg(org), nil
}

func (s *orgService) GetOrg(ctx context.Context, id uint, code string) (*contract.Org, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}

	var org *types.Organization
	var err error

	if id > 0 {
		org, err = db.GetOrgByID(ctx, s.db, id)
	} else if code != "" {
		org, err = db.GetOrgByCode(ctx, s.db, code)
	} else {
		return nil, errors.New("id or code is required")
	}

	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("org not found")
	}

	return convertToContractOrg(org), nil
}

func (s *orgService) UpdateOrg(ctx context.Context, id uint, req *contract.UpdateOrgRequest) (*contract.Org, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}
	if id == 0 {
		return nil, errors.New("id is required")
	}

	var org *types.Organization
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		org, err = db.GetOrgByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if org == nil {
			return errors.New("org not found")
		}

		if req.Name != nil {
			org.Name = strings.TrimSpace(*req.Name)
			if org.Name == "" {
				return errors.New("name cannot be empty")
			}
		}
		if req.Type != nil {
			org.Type = strings.TrimSpace(*req.Type)
		}
		if req.Status != nil {
			org.Status = strings.TrimSpace(*req.Status)
		}

		return db.UpdateOrg(ctx, tx, org)
	}); err != nil {
		return nil, err
	}

	return convertToContractOrg(org), nil
}

func (s *orgService) DeleteOrg(ctx context.Context, id uint) error {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return errors.New("user not authenticated")
	}
	if id == 0 {
		return errors.New("id is required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		org, err := db.GetOrgByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if org == nil {
			return errors.New("org not found")
		}
		return db.DeleteOrg(ctx, tx, id)
	})
}

func (s *orgService) ListOrgs(ctx context.Context, req *contract.ListOrgsRequest) (*contract.OrgList, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}
	req.Fill()

	opt := types.NewPageQuery(*caller, req.Offset, req.Limit)
	opt.ListAll = req.ListAll
	if req.Keyword != nil && *req.Keyword != "" {
		opt.AddFilter("keyword", *req.Keyword)
	}
	if req.Status != nil && *req.Status != "" {
		opt.AddFilter("status", *req.Status)
	}

	orgs, total, err := db.ListOrgs(ctx, s.db, opt)
	if err != nil {
		return nil, err
	}

	items := make([]contract.Org, 0, len(orgs))
	for _, org := range orgs {
		items = append(items, *convertToContractOrg(org))
	}
	return &contract.OrgList{
		Total:  total,
		Offset: req.Offset,
		Limit:  req.Limit,
		Items:  items,
	}, nil
}

func (s *orgService) CreateOrgMember(ctx context.Context, req *contract.CreateOrgMemberRequest) (*contract.OrgMember, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}
	if req.UserID == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.OrgID == 0 {
		return nil, errors.New("org_id is required")
	}

	user, err := db.GetUserByID(ctx, s.db, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	org, err := db.GetOrgByID(ctx, s.db, req.OrgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("org not found")
	}

	userOrg := &types.UserOrg{
		Uin:       req.UserID,
		UserID:    req.UserID,
		OrgID:     req.OrgID,
		IsDefault: req.IsDefault,
	}

	if err := db.CreateUserOrg(ctx, s.db, userOrg); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "Duplicate") {
			return nil, errors.New("org member already exists")
		}
		return nil, err
	}

	return s.enrichOrgMember(ctx, userOrg), nil
}

func (s *orgService) GetOrgMember(ctx context.Context, id uint, uin uint) (*contract.OrgMember, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}

	var userOrg *types.UserOrg
	var err error

	if id > 0 {
		userOrg, err = db.GetUserOrgByID(ctx, s.db, id)
	} else if uin > 0 {
		userOrg, err = db.GetUserOrgByUin(ctx, s.db, uin)
	} else {
		return nil, errors.New("id or uin is required")
	}

	if err != nil {
		return nil, err
	}
	if userOrg == nil {
		return nil, errors.New("org member not found")
	}

	return s.enrichOrgMember(ctx, userOrg), nil
}

func (s *orgService) UpdateOrgMember(ctx context.Context, id uint, req *contract.UpdateOrgMemberRequest) (*contract.OrgMember, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}
	if id == 0 {
		return nil, errors.New("id is required")
	}

	var userOrg *types.UserOrg
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		userOrg, err = db.GetUserOrgByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if userOrg == nil {
			return errors.New("org member not found")
		}

		if req.OrgID != nil {
			org, err := db.GetOrgByID(ctx, tx, *req.OrgID)
			if err != nil {
				return err
			}
			if org == nil {
				return errors.New("org not found")
			}
			userOrg.OrgID = *req.OrgID
		}
		if req.IsDefault != nil {
			userOrg.IsDefault = *req.IsDefault
		}

		return db.UpdateUserOrg(ctx, tx, userOrg)
	}); err != nil {
		return nil, err
	}

	return s.enrichOrgMember(ctx, userOrg), nil
}

func (s *orgService) DeleteOrgMember(ctx context.Context, id uint) error {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return errors.New("user not authenticated")
	}
	if id == 0 {
		return errors.New("id is required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userOrg, err := db.GetUserOrgByID(ctx, tx, id)
		if err != nil {
			return err
		}
		if userOrg == nil {
			return errors.New("org member not found")
		}
		return db.DeleteUserOrg(ctx, tx, id)
	})
}

func (s *orgService) ListOrgMembers(ctx context.Context, req *contract.ListOrgMembersRequest) (*contract.OrgMemberList, error) {
	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 {
		return nil, errors.New("user not authenticated")
	}
	req.Fill()

	opt := types.NewPageQuery(*caller, req.Offset, req.Limit)
	opt.ListAll = req.ListAll
	if req.OrgID != nil && *req.OrgID > 0 {
		opt.AddExactFilter("org_id", fmt.Sprintf("%d", *req.OrgID))
	}
	if req.UserID != nil && *req.UserID > 0 {
		opt.AddExactFilter("user_id", fmt.Sprintf("%d", *req.UserID))
	}

	userOrgs, total, err := db.ListUserOrgs(ctx, s.db, opt)
	if err != nil {
		return nil, err
	}

	items := make([]contract.OrgMember, 0, len(userOrgs))
	for _, uo := range userOrgs {
		items = append(items, *s.enrichOrgMember(ctx, uo))
	}
	return &contract.OrgMemberList{
		Total:  total,
		Offset: req.Offset,
		Limit:  req.Limit,
		Items:  items,
	}, nil
}

func (s *orgService) enrichOrgMember(ctx context.Context, uo *types.UserOrg) *contract.OrgMember {
	result := &contract.OrgMember{
		ID:        uo.ID,
		Uin:       uo.Uin,
		UserID:    uo.UserID,
		OrgID:     uo.OrgID,
		IsDefault: uo.IsDefault,
		CreatedAt: uo.CreatedAt,
		UpdatedAt: uo.UpdatedAt,
	}

	user, _ := db.GetUserByID(ctx, s.db, uo.UserID)
	if user != nil {
		result.UserName = user.Name
		result.UserLogin = user.GithubLogin
		result.AvatarURL = user.AvatarURL
	}

	org, _ := db.GetOrgByID(ctx, s.db, uo.OrgID)
	if org != nil {
		result.OrgName = org.Name
	}

	return result
}

func convertToContractOrg(org *types.Organization) *contract.Org {
	if org == nil {
		return nil
	}
	return &contract.Org{
		ID:        org.ID,
		Type:      org.Type,
		Code:      org.Code,
		Name:      org.Name,
		Status:    org.Status,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}
