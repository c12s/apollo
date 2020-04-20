package service

import (
	mPb "github.com/c12s/scheme/meridian"
	"google.golang.org/grpc"
	"log"
)

func NewMeridianClient(address string) mPb.MeridianServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to meridian service: %v", err)
	}

	return mPb.NewMeridianServiceClient(conn)
}
