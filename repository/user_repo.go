package repository

import (
	"apollo/model"
	"context"
)

type IUserRepo interface {
	CreateUser(ctx context.Context, req model.UserDTO) model.RegisterResp
	LoginUser(ctx context.Context, req model.LoginReq) model.LoginResp
}
