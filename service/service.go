package service

import (
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
	meridian   string
}

func (s *Server) List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "apollo.List")
	defer span.Finish()
	fmt.Println(span)

	token, err := helper.ExtractToken(ctx)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return nil, err
	}

	err = s.auth(ctx, listOpt(req, token))
	if err != nil {
		span.AddLog(&sg.KV{"auth error", err.Error()})
		return nil, err
	}

	_, err = s.checkNS(ctx, req.Extras["user"], req.Extras["namespace"])
	if err != nil {
		span.AddLog(&sg.KV{"check ns error", err.Error()})
		return nil, err
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

	token, err := helper.ExtractToken(ctx)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return nil, err
	}

	err = s.auth(ctx, mutateOpt(req, token))
	if err != nil {
		span.AddLog(&sg.KV{"auth error", err.Error()})
		return nil, err
	}

	_, err = s.checkNS(ctx, req.Mutate.UserId, req.Mutate.Namespace)
	if err != nil {
		span.AddLog(&sg.KV{"check ns error", err.Error()})
		return nil, err
	}

	rsp, err := s.db.Mutate(ctx, req)
	if err != nil {
		span.AddLog(&sg.KV{"roles mutate error", err.Error()})
		return nil, err
	}
	return rsp, nil
}

func (s *Server) GetToken(ctx context.Context, req *aPb.GetReq) (*aPb.GetResp, error) {
	return s.db.GetToken(ctx, req)
}

func (s *Server) Auth(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "apollo.auth")
	defer span.Finish()
	fmt.Println(span)

	rsp, err := s.db.Auth(ctx, req)
	if err != nil {
		span.AddLog(&sg.KV{"auth error", err.Error()})
		return nil, err
	}

	if v, ok := rsp.Data["intent"]; ok && v == "register" && rsp.Value {
		token, err := helper.ExtractToken(ctx)
		if err != nil {
			span.AddLog(&sg.KV{"token error", err.Error()})
			return nil, err
		}
		err = s.createDefaultNamespace(
			helper.AppendToken(
				sg.NewTracedGRPCContext(ctx, span),
				token,
			),
			req.Data["username"])
		if err != nil {
			return nil, err
		}
	}

	return rsp, nil
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
		meridian:   conf.Meridian,
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
