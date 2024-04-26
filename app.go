package main

import (
	"apollo/configs"
	"apollo/model"
	"apollo/model/db"
	"apollo/proto1"
	"apollo/server"
	"apollo/service"
	"apollo/vault"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	config                    configs.Config
	grpcServer                *grpc.Server
	authServiceServer         proto1.AuthServiceServer
	authService               *service.AuthService
	vaultService              *vault.VaultClientService
	authRepo                  model.UserRepo
	cm                        *db.CassandraManager
	shutdownProcesses         []func()
	gracefulShutdownProcesses []func(wg *sync.WaitGroup)
}

func NewAppWithConfig(config configs.Config) (*app, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	return &app{
		config:                    config,
		shutdownProcesses:         make([]func(), 0),
		gracefulShutdownProcesses: make([]func(wg *sync.WaitGroup), 0),
	}, nil
}

func (a *app) Start() error {
	a.init()

	return a.startGrpcServer()
}

func (a *app) GracefulStop(ctx context.Context) {
	// call all shutdown processes after a timeout or graceful shutdown processes completion
	defer a.shutdown()

	// wait for all graceful shutdown processes to complete
	wg := &sync.WaitGroup{}
	wg.Add(len(a.gracefulShutdownProcesses))

	for _, gracefulShutdownProcess := range a.gracefulShutdownProcesses {
		go gracefulShutdownProcess(wg)
	}

	// notify when graceful shutdown processes are done
	gracefulShutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		gracefulShutdownDone <- struct{}{}
	}()

	// wait for graceful shutdown processes to complete or for ctx timeout
	select {
	case <-ctx.Done():
		log.Println("ctx timeout ... shutting down")
	case <-gracefulShutdownDone:
		log.Println("app gracefully stopped")
	}
}

func (a *app) init() {
	manager := db.NewCassandraManager()
	a.cm = manager

	a.initUserRepo(a.cm)

	a.initVaultClientService()
	a.initAuthService()

	a.initAuthServiceServer()
	a.initGrpcServer()
}

func (a *app) initGrpcServer() {

	if a.authServiceServer == nil {
		log.Fatalln("eval grpc server is nil")
	}
	s := grpc.NewServer()
	proto1.RegisterAuthServiceServer(s, a.authServiceServer)
	reflection.Register(s)
	a.grpcServer = s
}

func (a *app) initAuthServiceServer() {
	if a.authService == nil {
		log.Fatalln("Auth service is nil")
	}
	server, err := server.NewAuthServiceServer(*a.authService)
	if err != nil {
		log.Fatalln(err)
	}
	a.authServiceServer = server
}

func (a *app) initVaultClientService() {
	vaultService, err := vault.NewVaultClientService()
	if err != nil {
		log.Fatalln(err)
	}
	a.vaultService = vaultService
}

func (a *app) initAuthService() {
	if a.authRepo == nil {
		log.Fatalln("auth repo is nil")
	}
	authService, err := service.NewAuthService(a.authRepo, a.vaultService)
	if err != nil {
		log.Fatalln(err)
	}
	a.authService = authService
}

func (a *app) initUserRepo(manager *db.CassandraManager) {
	a.authRepo = db.NewUserRepo(manager)
}

func (a *app) startGrpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.config.Server().Port()))
	if err != nil {
		return err
	}
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	a.gracefulShutdownProcesses = append(a.gracefulShutdownProcesses, func(wg *sync.WaitGroup) {
		a.grpcServer.GracefulStop()
		log.Println("apollo server gracefully stopped")
		wg.Done()
	})
	return nil
}

func (a *app) shutdown() {
	for _, shutdownProcess := range a.shutdownProcesses {
		shutdownProcess()
	}
}
