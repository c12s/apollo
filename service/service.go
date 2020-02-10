package service

import (
	"errors"
	"fmt"
	"github.com/c12s/apollo/helper"
	"github.com/c12s/apollo/model"
	"github.com/c12s/apollo/storage"
	aPb "github.com/c12s/scheme/apollo"
	cPb "github.com/c12s/scheme/celestial"
	sg "github.com/c12s/stellar-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type Server struct {
	instrument map[string]string
	db         storage.DB
}

func (s *Server) List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "apollo.List")
	defer span.Finish()
	fmt.Println(span)

	resp, err := s.Auth(ctx,
		&aPb.AuthOpt{
			Data: map[string]string{"intent": "auth"},
		},
	)

	if err != nil {
		span.AddLog(&sg.KV{"apollo resp error", err.Error()})
		return nil, err
	}

	if !resp.Value {
		span.AddLog(&sg.KV{"apollo.auth value", resp.Data["message"]})
		return nil, errors.New(resp.Data["message"])
	}

	rsp, err := s.db.List(ctx, req)
	if err != nil {
		span.AddLog(&sg.KV{"roles list error", err.Error()})
		return nil, err
	}
	return rsp, nil
}

func (s *Server) Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "apollo.Mutate")
	defer span.Finish()
	fmt.Println(span)

	resp, err := s.Auth(
		ctx,
		&aPb.AuthOpt{
			Data: map[string]string{"intent": "auth"},
		},
	)

	if err != nil {
		span.AddLog(&sg.KV{"apollo resp error", err.Error()})
		return nil, err
	}

	if !resp.Value {
		span.AddLog(&sg.KV{"apollo.auth value", resp.Data["message"]})
		return nil, errors.New(resp.Data["message"])
	}

	rsp, err := s.db.Mutate(ctx, req)
	if err != nil {
		span.AddLog(&sg.KV{"roles mutate error", err.Error()})
		return nil, err
	}
	return rsp, nil
}

func (s *Server) GetToken(ctx context.Context, req *aPb.GetReq) (*aPb.GetResp, error) {
	return &aPb.GetResp{"myroot"}, nil //TODO: This is test only, dummy value!
}

func (s *Server) Auth(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "apollo.auth")
	defer span.Finish()
	fmt.Println(span)

	if req.Data["intent"] == "login" {
		span.AddLog(&sg.KV{"apollo auth value", "received intent login"})

		return &aPb.AuthResp{
			Value: true,
			Data: map[string]string{
				"token":   "some_random_token",
				"message": "logged in",
			},
		}, nil
	} else if req.Data["intent"] == "auth" {
		token, err := helper.ExtractToken(ctx)
		if err != nil {
			fmt.Println(err.Error())
			span.AddLog(&sg.KV{"token error", err.Error()})
			return nil, err
		}

		fmt.Println("TOKEN: ", token)

	} else {
		fmt.Println("RECEIVED INTENT: ", req.Data, req.Extras)
	}

	return &aPb.AuthResp{
		Value: true,
		Data: map[string]string{
			"message": "You do not have access for that action",
		},
	}, nil
}

func Run(db storage.DB, conf *model.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lis, err := net.Listen("tcp", conf.Address)
	if err != nil {
		log.Fatalf("failed to initializa TCP listen: %v", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	apolloServer := &Server{
		instrument: conf.InstrumentConf,
		db:         db,
	}

	n, err := sg.NewCollector(apolloServer.instrument["address"], apolloServer.instrument["stopic"])
	if err != nil {
		fmt.Println(err)
		return
	}
	c, err := sg.InitCollector(apolloServer.instrument["location"], n)
	if err != nil {
		fmt.Println(err)
		return
	}
	go c.Start(ctx, 15*time.Second)

	fmt.Println("ApolloService RPC Started")
	aPb.RegisterApolloServiceServer(server, apolloServer)
	server.Serve(lis)
}
