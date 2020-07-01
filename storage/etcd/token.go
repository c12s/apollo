package etcd

import (
	"context"
	"errors"
	aPb "github.com/c12s/scheme/apollo"
	"github.com/golang/protobuf/proto"
)

func (db DB) create(ctx context.Context, key string, data *aPb.AuthOpt) (*aPb.GetResp, error) {
	token, err := db.secrets.CreateToken(ctx, data.Data)
	if err != nil {
		return nil, err
	}
	data.Data["token"] = token
	err = db.update(ctx, data, key)
	if err != nil {
		return nil, err
	}
	return &aPb.GetResp{Token: token}, nil
}

func (db DB) GetToken(ctx context.Context, req *aPb.GetReq) (*aPb.GetResp, error) {
	key := usersKeyspace(req.User, "default") //TODO:THIS MUST BE FIXED ONCE PROTOBIF IS!!
	value, err := db.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	for _, item := range value.Kvs {
		data := &aPb.AuthOpt{}
		err = proto.Unmarshal(item.Value, data)
		if err != nil {
			return nil, err
		}

		if _, ok := data.Data["token"]; !ok {
			return db.create(ctx, key, data)
		}

		secret, err := db.secrets.GetToken(ctx, data.Data["token"])
		if err != nil {
			newToken, err := db.secrets.CreateToken(ctx, data.Data)
			if err != nil {
				return nil, err
			}
			if data.Data["token"] != newToken {
				data.Data["token"] = newToken
				err = db.update(ctx, data, key)
				if err != nil {
					return nil, err
				}
			}
			return &aPb.GetResp{Token: newToken}, nil
		}
		return &aPb.GetResp{Token: secret["id"].(string)}, nil
	}
	return nil, errors.New("Token or user do not exists")
}
