package model

import (
	"context"
)

type UserRepo interface {
	CreateUser(ctx context.Context, req User) RegisterResp
	LoginUser(ctx context.Context, req LoginReq) LoginResp
	GetUserPermissions(ctx context.Context, org_id string, user_id string) []string
}




