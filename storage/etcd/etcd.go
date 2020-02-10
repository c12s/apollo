package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/apollo/model"
	aPb "github.com/c12s/scheme/apollo"
	cPb "github.com/c12s/scheme/celestial"
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"time"
)

type DB struct {
	kv     clientv3.KV
	client *clientv3.Client
}

func New(conf *model.Config, timeout time.Duration) (*DB, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: timeout,
		Endpoints:   conf.DB,
	})

	if err != nil {
		return nil, err
	}

	return &DB{
		kv:     clientv3.NewKV(cli),
		client: cli,
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
