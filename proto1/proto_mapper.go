package proto1

import (
	"apollo/model"
)

func UserToModel(req *User) (*model.User, error) {
	org := req.Org

	if org == "" {
		org = req.Username + "_default"
	}
	return &model.User{
		Name:       		 req.Name,
		Surname:         	 req.Surname,
		Email: 				 req.Email,
		Password:            req.Password,
		Org:            	 org,
		Username:            req.Username,
	}, nil
}

func LoginToModel(req *LoginReq) (*model.LoginReq, error) {
	return &model.LoginReq{
		Username: 			req.Username,
		Password:           req.Password,
	}, nil
}

func TokenToModel(req *Token) (*model.Token, error) {
	return &model.Token{
		Token: 			req.Token,
	}, nil
}

func JwtToModel(req *InternalToken) (*model.Token, error) {
	return &model.Token{
		Token: 			req.Jwt,
	}, nil
}
