package repository

import (
	"apollo/model"
	"apollo/model/dao"
	"apollo/proto1"
	"context"
	"errors"
	"log"
)

type UserRepo struct {
	userHandler *dao.UserHandler
	orgHandler  *dao.OrgHandler
	permHandler *dao.PermHandler
}

func NewUserRepo(userHandler *dao.UserHandler, orgHandler *dao.OrgHandler, permHandler *dao.PermHandler) IUserRepo {
	return UserRepo{
		userHandler: userHandler,
		orgHandler:  orgHandler,
		permHandler: permHandler,
	}
}

func (store UserRepo) CreateUser(ctx context.Context, req model.UserDTO) model.RegisterResp {
	user, _ := proto1.UserToModel(&req)
	_, err := store.orgHandler.FindOrgByName(ctx, req.Org)

	if err == nil {
		return model.RegisterResp{User: model.UserDTO{}, Error: errors.New("organization already exists")}
	}

	_, err = store.userHandler.FindUserByUsername(ctx, user.Username)
	if err == nil {
		return model.RegisterResp{User: model.UserDTO{}, Error: errors.New("user already exists")}
	}

	_, err = store.userHandler.FindUserByEmail(ctx, user.Email)
	if err == nil {
		return model.RegisterResp{User: model.UserDTO{}, Error: errors.New("user already exists")}
	}

	err = store.userHandler.InsertUser(ctx, user)
	if err != nil {
		log.Printf("Registration of user failed")
		return model.RegisterResp{User: model.UserDTO{}, Error: err}
	}

	permissions, err := store.permHandler.GetAllPerms(ctx)
	if err != nil {
		log.Printf("GetUserPermissions failed")
		return model.RegisterResp{User: model.UserDTO{}, Error: err}
	}

	err = store.orgHandler.InsertOrg(ctx, &model.Org{
		Name:        req.Org,
		Owner:       req.Email,
		Members:     nil,         // user is owner in his org by default
		Permissions: permissions, // user gets all available permissions
	})
	if err != nil {
		log.Printf("Insertion of new org failed. %v", err)
		return model.RegisterResp{User: model.UserDTO{}, Error: err}
	}

	return model.RegisterResp{User: model.UserDTO{
		Name:        req.Name,
		Surname:     req.Surname,
		Org:         req.Org,
		Permissions: permissions,
		Username:    req.Username,
		Email:       req.Email,
	}, Error: nil}

}

func (store UserRepo) LoginUser(ctx context.Context, req model.LoginReq) model.LoginResp {
	return model.LoginResp{Token: "", Error: errors.New("invalid mapping")}
}
