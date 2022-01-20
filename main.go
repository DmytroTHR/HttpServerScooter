package main

import (
	configs "HttpServer/config"
	protoProblem "HttpServer/proto/protoProblem"
	protoUser "HttpServer/proto/protoUser"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type servicesGRPC struct {
	userService    protoUser.UserServiceClient
	problemService protoProblem.ProblemServiceClient
}

func getProblemService() *grpc.ClientConn {
	problemGRPCServer := net.JoinHostPort(configs.PROBLEMS_SERVICE, configs.PROBLEMS_GRPC_PORT)
	problemCred, err := credentials.NewClientTLSFromFile(configs.PROBLEMS_CERTIFICATE, "")
	if err != nil {
		log.Panicf("%s: unable to get TLS certificate - %v", problemGRPCServer, err)
	}
	problemConnection, err := grpc.Dial(problemGRPCServer, grpc.WithTransportCredentials(problemCred))
	if err != nil {
		log.Panicf("%s: unable to set grpc connection - %v", problemGRPCServer, err)
	}
	return problemConnection
}

func getUserService() *grpc.ClientConn {
	userGRPCServer := net.JoinHostPort(configs.USER_SERVICE, configs.USERS_GRPC_PORT)
	userConnection, err := grpc.Dial(userGRPCServer, grpc.WithInsecure())
	if err != nil {
		log.Panicf("%s: unable to set grpc connection - %v", userGRPCServer, err)
	}
	return userConnection
}

var clientsGRPC = &servicesGRPC{}

func main() {

	problemConnection := getProblemService()
	defer problemConnection.Close()
	clientsGRPC.problemService = protoProblem.NewProblemServiceClient(problemConnection)

	userConnection := getUserService()
	defer userConnection.Close()
	clientsGRPC.userService = protoUser.NewUserServiceClient(userConnection)

	httpServer := &http.Server{
		Handler: &requestHandler{},
		Addr:    ":8888",
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		log.Println("Server shutting down..")
		close(idleConnsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}