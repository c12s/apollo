package db

import (
	"apollo/model"
	"log"
	"errors"
	"strings"
	"context"
)

type UserRepo struct {
	manager *CassandraManager
}

func NewUserRepo(manager *CassandraManager) model.UserRepo {
	return UserRepo{
		manager: manager,
	}
}

func (store UserRepo) CreateUser(ctx context.Context, req model.User) model.RegisterResp {
	foundOrg, err := store.manager.FindOrgByName(ctx, req.Org)

	if err != nil {
		log.Printf("The organization with name %s doesn't exist. The default org is going to be created", req.Org)
		orgName := req.Org
		if strings.TrimSpace(req.Org) == "" {
			orgName = req.Username + "_default"
		}
		
		newOrg := model.Org{
			Name: orgName,
		}
		foundOrg, err = store.manager.InsertOrg(ctx, newOrg)
		if err != nil {
			log.Printf("Insertion of new org failed.")
			return model.RegisterResp{User: model.User{}, Error: err}
		}
	} else {
		return model.RegisterResp{User: model.User{}, Error: errors.New("Organization already exists.")}
	}

	userId, err := store.manager.InsertUser(ctx, req)
	if err != nil {
		log.Printf("Registration of user failed")
		return model.RegisterResp{User: model.User{}, Error: err}
	}

	// connect org and user
	_, err = store.manager.CreateOrgUser(foundOrg.Id, userId, true)
	
	if err != nil {
		log.Printf("User - org relationship failed")
		return model.RegisterResp{User: model.User{}, Error: err}
	}

	permissions, err := store.manager.GetUserPermissions(foundOrg.Id, userId)

	if err != nil {
		log.Printf("GetUserPermissions failed")
		return model.RegisterResp{User: model.User{}, Error: err}
	}

	return model.RegisterResp{User: model.User{
		Id: userId,
		Name: req.Name,
		Surname: req.Surname,
		Org: req.Org,
		Permissions: permissions,
		Username: req.Username,
		Email: req.Email,
	}, Error: nil}		

}

func (store UserRepo) LoginUser(ctx context.Context, req model.LoginReq) model.LoginResp {
	return model.LoginResp{Token: "", Error: errors.New("Invalid mapping")}		
}

func (store UserRepo) GetUserPermissions(ctx context.Context, org_id string, user_id string) []string {
	permissions, err := store.manager.GetUserPermissions(org_id, user_id)

	if err != nil {
		log.Println("User permissions not found")
	}

	return permissions
}

