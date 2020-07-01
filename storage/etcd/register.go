package etcd

import (
	"context"
	"errors"
	aPb "github.com/c12s/scheme/apollo"
	"github.com/golang/protobuf/proto"
	"time"
)

func (db *DB) register(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	key := usersKeyspace(req.Data["username"], "default") // for now all are in default ns
	resp, err := db.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) > 0 {
		return nil, errors.New("username exists")
	}

	pass, err := hashPassword(req.Data["password"])
	if err != nil {
		return nil, err
	}
	req.Data["password"] = pass
	token, err := db.secrets.CreateToken(ctx, map[string]string{
		"username":  req.Data["username"],
		"password":  pass,
		"namespace": "default",
	})
	if err != nil {
		return nil, err
	}
	req.Data["token"] = token

	nsData, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = db.kv.Put(ctx, key, string(nsData))
	if err != nil {
		return nil, err
	}

	_ = db.defaultRoles(ctx, req.Data["username"])

	return &aPb.AuthResp{
		Value: true,
		Data: map[string]string{
			"message": "User created.",
			"intent":  "register",
		},
	}, nil
}

func (db *DB) update(ctx context.Context, req *aPb.AuthOpt, key string) error {
	nsData, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	_, err = db.kv.Put(ctx, key, string(nsData))
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) defaultRoles(ctx context.Context, username string) error {
	key := userKeyspace(username, "default")
	verbs := []string{"mutate", "list"}
	for _, res := range []string{"configs", "secrets", "namespaces", "actions", "roles", "topology"} {
		resKey := resourceKeyspace(key, res)
		rs, err := proto.Marshal(&aPb.ACL{
			Token:   verbs,
			Created: time.Now().Unix(),
		})
		_, err = db.kv.Put(ctx, resKey, string(rs))
		if err != nil {
			return err
		}
	}
	return nil
}
