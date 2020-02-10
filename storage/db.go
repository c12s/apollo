package storage

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type DB interface {
	List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error)
	Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error)
}
