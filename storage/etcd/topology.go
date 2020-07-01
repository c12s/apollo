package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	aPb "github.com/c12s/scheme/apollo"
	"github.com/golang/protobuf/proto"
	"time"
)

func (db *DB) topology(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	if req.Data["action"] == "list" {
		return db.listTopology(ctx, req)
	} else if req.Data["action"] == "mutate" {
		return db.mutateTopology(ctx, req)
	} else {
		return nil, errors.New("Invalid action received.")
	}
}

func (db *DB) listTopology(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	key := userKeyspace(req.Extras["user"].Data[0], "default") //TODO: For now all topologies goes in default namespace
	resKey := resourceKeyspace(key, "toplogy")

	value, err := db.cache.Get(resKey)
	if err == nil {
		var verbs []string
		err := json.Unmarshal([]byte(value.(string)), &verbs)
		if err == nil {
			fmt.Println("CACHE HIT")
			for _, verb := range verbs {
				if verb == "*" || verb == "list" {
					return &aPb.AuthResp{Value: true}, nil
				}
			}
		}
	}

	resp, err := db.client.Get(ctx, resKey)
	if err != nil {
		return nil, err
	}

	for _, item := range resp.Kvs {
		acl := &aPb.ACL{}
		err = proto.Unmarshal(item.Value, acl)
		if err != nil {
			return nil, err
		}
		db.cache.Put(resKey, acl.Token, 10*time.Minute)
		fmt.Println("CACHED NOW")

		for _, verb := range acl.Token {
			if verb == "*" || verb == "list" {
				return &aPb.AuthResp{Value: true}, nil
			}
		}
	}

	return &aPb.AuthResp{
		Value: false,
		Data: map[string]string{
			"message": "You do not have access for that action",
		},
	}, nil
}

func (db *DB) mutateTopology(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	key := userKeyspace(req.Data["user"], "default") //TODO: for now all topology goes in default namespace
	resKey := resourceKeyspace(key, "topology")

	value, err := db.cache.Get(resKey)
	if err == nil {
		var verbs []string
		err := json.Unmarshal([]byte(value.(string)), &verbs)
		if err == nil {
			fmt.Println("CACHE HIT")
			for _, verb := range verbs {
				if verb == "*" || verb == "mutate" {
					return &aPb.AuthResp{Value: true}, nil
				}
			}
		}
	}

	resp, err := db.client.Get(ctx, resKey)
	if err != nil {
		return nil, err
	}

	for _, item := range resp.Kvs {
		acl := &aPb.ACL{}
		err = proto.Unmarshal(item.Value, acl)
		if err != nil {
			return nil, err
		}
		db.cache.Put(resKey, acl.Token, 10*time.Minute)
		fmt.Println("CACHED NOW")

		for _, verb := range acl.Token {
			if verb == "*" || verb == "mutate" {
				return &aPb.AuthResp{Value: true}, nil
			}
		}
	}

	return &aPb.AuthResp{
		Value: false,
		Data: map[string]string{
			"message": "You do not have access for that action",
		},
	}, nil
}
