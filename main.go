package main

import (
	configs "HttpServer/config"
	protoProblem "HttpServer/proto/protoProblem"
	protoUser "HttpServer/proto/protoUser"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

type requestHandler struct {
}

func (rh *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		queryStr, err := bodyToString(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Unable to convert body to string: %v", err)
			return
		}

		workWithRequest(queryStr, w)
	} else {
		fmt.Fprintln(w, "Server is working...")
	}
}

func bodyToString(closer io.ReadCloser) (string, error) {
	bodyBytes, err := ioutil.ReadAll(closer)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func workWithRequest(query string, w http.ResponseWriter) {
	var request struct {
		Method     string
		Parameters interface{}
	}
	decoder := json.NewDecoder(strings.NewReader(query))
	err := decoder.Decode(&request)
	if err != nil {
		log.Panicf("error decoding request: %v", err)
	}

	ctx := context.Background()
	var result interface{}
	switch request.Method {
	case "CreateUser":
		userToCreate := getUserDataFromRequestParams(request.Parameters)
		result, err = clientsGRPC.userService.CreateUser(ctx, userToCreate)
		if err != nil {
			fmt.Fprintf(w, "CreateUser error: %v", err)
			return
		}
	case "GetUser":
		mapParams, ok := request.Parameters.(map[string]interface{})
		if !ok {
			fmt.Fprintf(w, "GetUser problem: cannot map params")
			return
		}
		idUser, ok := mapParams["id"]
		if !ok {
			fmt.Fprintf(w, "GetUser problem: cannot get ID")
			return
		}
		idConverted, err := strconv.Atoi(idUser.(string))
		if err != nil {
			fmt.Fprintf(w, "GetUser problem: cannot convert ID to int64 - %v", err)
			return
		}
		result, err = clientsGRPC.userService.GetUserByID(ctx, &protoUser.User{Id: int64(idConverted)})
		if err != nil {
			fmt.Fprintf(w, "GetUser error: %v", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getUserDataFromRequestParams(params interface{}) *protoUser.User {
	resUser := &protoUser.User{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resUser
	}

	resUser.Name = mapParam["name"].(string)
	resUser.Surname = mapParam["surname"].(string)
	resUser.Email = mapParam["email"].(string)
	resUser.Password = mapParam["password"].(string)
	resUser.Role = &protoUser.Role{Id: 2}
	return resUser
}
