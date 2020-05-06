package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/c12s/apollo/helper"
	"github.com/c12s/apollo/model"
	"github.com/c12s/apollo/storage"
	aPb "github.com/c12s/scheme/apollo"
	cPb "github.com/c12s/scheme/celestial"
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"time"
)

type DB struct {
	kv      clientv3.KV
	client  *clientv3.Client
	cache   storage.Cacher
	secrets storage.Secrets
}

func New(conf *model.Config, cache storage.Cacher, secrets storage.Secrets, timeout time.Duration) (*DB, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: timeout,
		Endpoints:   conf.DB,
	})

	if err != nil {
		return nil, err
	}

	return &DB{
		kv:      clientv3.NewKV(cli),
		client:  cli,
		cache:   cache,
		secrets: secrets,
	}, nil
}

func (db *DB) Close() { db.client.Close() }

func (r *DB) List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	users := split(req.Extras["users"])
	resources := split(req.Extras["resources"])
	namespaces := split(req.Extras["namespaces"])

	if len(namespaces) == 0 {
		namespaces = append(namespaces, "default")
	}

	rez := map[string]string{}
	for _, user := range users {
		for _, namespace := range namespaces {
			key := userKeyspace(user, namespace)

			chspan := span.Child("etcd.get")
			if len(resources) > 0 {
				for _, resource := range resources {
					resp, err := r.client.Get(ctx, resourceKeyspace(key, resource))
					if err != nil {
						chspan.AddLog(&sg.KV{"etcd get error", err.Error()})
						return nil, err
					}

					for _, item := range resp.Kvs {
						dt := &aPb.ACL{}
						err = proto.Unmarshal(item.Value, dt)
						if err != nil {
							span.AddLog(&sg.KV{"unmarshall etcd get error", err.Error()})
							return nil, err
						}

						rezKey := join(":", []string{user, namespace, resource, toString(dt.Created)})
						rez[rezKey] = join(",", dt.Token)
					}
				}
			} else {
				resp, err := r.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
				if err != nil {
					chspan.AddLog(&sg.KV{"etcd get error", err.Error()})
					return nil, err
				}

				for _, item := range resp.Kvs {
					dt := &aPb.ACL{}
					err = proto.Unmarshal(item.Value, dt)
					if err != nil {
						span.AddLog(&sg.KV{"unmarshall etcd get error", err.Error()})
						return nil, err
					}

					k := ssplit(string(item.Key), "/")
					rezKey := join(":", []string{user, namespace, k[len(k)-1], toString(dt.Created)})
					rez[rezKey] = join(",", dt.Token)
				}
			}
			go chspan.Finish()
		}
	}
	return &cPb.ListResp{Extras: rez}, nil
}

func (r DB) Init(ctx context.Context, userid, namespace string) {
	key := userKeyspace(userid, namespace)
	res := []string{"actions", "secrets", "configs", "roles", "namespaces"}
	for _, resource := range res {
		rs, err := proto.Marshal(&aPb.ACL{
			Token:   []string{"*"},
			Created: 0,
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		resKey := resourceKeyspace(key, resource)
		_, err = r.kv.Put(ctx, resKey, string(rs))
		if err != nil {
			fmt.Println(err.Error())
			continue
		} else {
			fmt.Println("{{EVENT}} Added resource", resource)
		}
	}
}

func (r *DB) Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	key := userKeyspace(req.Mutate.Extras["user"], req.Mutate.Namespace)
	resources := split(req.Mutate.Extras["resources"])
	fmt.Println(req)

	chspan1 := span.Child("etcd.delete key")
	_, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		chspan1.AddLog(&sg.KV{"etcd.delete key error", err.Error()})
		return nil, err
	}
	chspan1.Finish()

	for _, resource := range resources {
		rs, err := proto.Marshal(&aPb.ACL{
			Token:   split(req.Mutate.Extras["verbs"]),
			Created: req.Mutate.Timestamp,
		})
		if err != nil {
			span.AddLog(&sg.KV{"marshaling error", err.Error()})
			return nil, err
		}

		resKey := resourceKeyspace(key, resource)
		chspan2 := span.Child("etcd.put")
		_, err = r.kv.Put(ctx, resKey, string(rs))
		if err != nil {
			chspan2.AddLog(&sg.KV{"etcd.put error", err.Error()})
			return nil, err
		}
		chspan2.Finish()
	}

	span.AddLog(&sg.KV{"role addition", "Role added."})
	return &cPb.MutateResp{"Role added."}, nil
}

func (db *DB) Auth(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "db.auth")
	defer span.Finish()
	fmt.Println(span)

	if req.Data["intent"] == "login" {
		span.AddLog(&sg.KV{"apollo auth value", "received intent login"})
		return db.login(ctx, req)
	} else if req.Data["intent"] == "auth" {
		token, err := helper.ExtractToken(ctx)
		if err != nil {
			fmt.Println(err.Error())
			span.AddLog(&sg.KV{"token error", err.Error()})
			return nil, err
		}

		fmt.Println("TOKEN: ", token)
		fmt.Println("RECEIVED INTENT AUTH: ", req.Data, req.Extras)

		switch req.Data["kind"] {
		case "roles":
			return db.roles(ctx, req)
		case "configs":
			return db.configs(ctx, req)
		case "secrets":
			return db.secret(ctx, req)
		case "actions":
			return db.actions(ctx, req)
		case "namespaces":
			return db.namespaces(ctx, req)
		}
		return nil, errors.New("Invalid kind")
	} else if req.Data["intent"] == "register" {
		return db.register(ctx, req)
	} else {
		fmt.Println("RECEIVED INTENT ELSE: ", req.Data, req.Extras)
	}

	return &aPb.AuthResp{
		Value: true,
		Data: map[string]string{
			"message": "You do not have access for that action",
		},
	}, nil
}
