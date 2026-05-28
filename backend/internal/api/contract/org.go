package contract

import "context"

type OrgService interface {
	CreateOrg(ctx context.Context, req *CreateOrgRequest) (*Org, error)
	GetOrg(ctx context.Context, id uint, code string) (*Org, error)
	UpdateOrg(ctx context.Context, id uint, req *UpdateOrgRequest) (*Org, error)
	DeleteOrg(ctx context.Context, id uint) error
	ListOrgs(ctx context.Context, req *ListOrgsRequest) (*OrgList, error)

	CreateOrgMember(ctx context.Context, req *CreateOrgMemberRequest) (*OrgMember, error)
	GetOrgMember(ctx context.Context, id uint, uin uint) (*OrgMember, error)
	UpdateOrgMember(ctx context.Context, id uint, req *UpdateOrgMemberRequest) (*OrgMember, error)
	DeleteOrgMember(ctx context.Context, id uint) error
	ListOrgMembers(ctx context.Context, req *ListOrgMembersRequest) (*OrgMemberList, error)
}
