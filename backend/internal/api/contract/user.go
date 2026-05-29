package contract

import "context"

type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserInfo, error)
	GetUser(ctx context.Context, publicID string, githubLogin string) (*UserInfo, error)
	UpdateUser(ctx context.Context, publicID string, req *UpdateUserRequest) (*UserInfo, error)
	DeleteUser(ctx context.Context, publicID string) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (*UserList, error)
}
