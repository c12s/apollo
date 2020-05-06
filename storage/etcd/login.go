package etcd

import (
	"context"
	"errors"
	aPb "github.com/c12s/scheme/apollo"
	"github.com/golang/protobuf/proto"
)

func (db *DB) login(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	key := usersKeyspace(req.Data["username"], "default") // for now all are in default ns
	resp, err := db.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errors.New("username do not exists")
	}
	for _, item := range resp.Kvs {
		apt := &aPb.AuthOpt{}
		err = proto.Unmarshal(item.Value, apt)
		if err != nil {
			return nil, err
		}
		if checkPasswordHash(req.Data["password"], apt.Data["password"]) {
			return &aPb.AuthResp{
				Value: true,
				Data: map[string]string{
					"token":   db.generateToken(),
					"message": "You are no logged in",
				},
			}, nil
		}
	}

	return &aPb.AuthResp{
		Value: false,
		Data: map[string]string{
			"message": "invalid login credentials",
		},
	}, nil
}

func (db DB) generateToken() string {
	return "some_random_token"
}
