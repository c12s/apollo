package service

import (
	"fmt"
	aPb "github.com/c12s/scheme/apollo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct{}

func (s *Server) GetToken(ctx context.Context, req *aPb.GetReq) (*aPb.GetResp, error) {
	return &aPb.GetResp{"myroot"}, nil //TODO: This is test only, dummy value!
}

func (s *Server) Auth(ctx context.Context, req *aPb.AuthOpt) (*aPb.AuthResp, error) {
	fmt.Println(req)

	return &aPb.AuthResp{
		Value: true,
		Data: map[string]string{
			"token": "some_random_token",
		},
	}, nil
}

func Run(ctx context.Context, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to initializa TCP listen: %v", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	apolloServer := &Server{}

	fmt.Println("ApolloService RPC Started")
	aPb.RegisterApolloServiceServer(server, apolloServer)
	server.Serve(lis)
}
