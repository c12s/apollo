package server

import (
	"context"
	"fmt"
	"apollo/proto1"
	"apollo/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServiceServer struct {
	service service.AuthService
	proto1.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(service service.AuthService) (proto1.AuthServiceServer, error) {
	return &AuthServiceServer{
		service: service,
	}, nil
}

func (o *AuthServiceServer) Authorize(ctx context.Context, req *proto1.AuthorizationReq) (*proto1.AuthorizationResp, error) {
	return &proto1.AuthorizationResp{Authorized: true}, nil
}

func (o *AuthServiceServer) RegisterUser(ctx context.Context, req *proto1.User) (*proto1.RegResp, error) {
	user, err := proto1.UserToModel(req)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%s", err))
	}

	resp := o.service.RegisterUser(ctx, *user)

	if resp.Error != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%s", resp.Error))
	}

	return &proto1.RegResp{User: &proto1.RegisteredUser{
		Id:      resp.User.Id,
		Name:    resp.User.Name,
		Surname: resp.User.Surname,
		Email:   resp.User.Email}}, nil
}

func (o *AuthServiceServer) LoginUser(ctx context.Context, req *proto1.LoginReq) (*proto1.LoginResp, error) {
	user, err := proto1.LoginToModel(req)

	if err != nil {
		return nil, status.Error(codes.Internal, "Error in login request")
	}

	resp := o.service.LoginUser(*user)

	if resp.Error != nil {
		return nil, status.Error(codes.Internal, "Invalid username and/or password")
	}

	return &proto1.LoginResp{Token: resp.Token}, nil
}

func (o *AuthServiceServer) VerifyToken(ctx context.Context, req *proto1.Token) (*proto1.VerifyResp, error) {
	token, err := proto1.TokenToModel(req)

	if err != nil {
		return nil, status.Error(codes.Internal, "Error in token request")
	}

	resp, username := o.service.VerifyToken(*token)

	if !resp.Verified {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	return &proto1.VerifyResp{Token: &proto1.InternalToken{Verified: resp.Verified,
		Jwt: resp.Jwt,
	},
		Username: username}, nil
}

func (o *AuthServiceServer) DecodeJwt(ctx context.Context, req *proto1.Token) (*proto1.DecodedJwtResp, error) {
	token, err := proto1.TokenToModel(req)

	if err != nil {
		return nil, status.Error(codes.Internal, "Error in decoding jwt")
	}

	resp := o.service.DecodeJwt(*token)
	return &proto1.DecodedJwtResp{Permissions: resp}, nil
}
