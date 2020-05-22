package storage

import (
	"context"
	aPb "github.com/c12s/scheme/apollo"
	cPb "github.com/c12s/scheme/celestial"
)

type DB interface {
	List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error)
	Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error)
	Auth(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error)
	Init(ctx context.Context, userid, namespace string)
	GetToken(ctx context.Context, req *aPb.GetReq) (*aPb.GetResp, error)
}
