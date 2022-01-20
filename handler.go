package main

import (
	protoProblem "HttpServer/proto/protoProblem"
	protoUser "HttpServer/proto/protoUser"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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
		userToCreate := getUserData(request.Parameters)
		result, err = clientsGRPC.userService.CreateUser(ctx, userToCreate)

	case "GetUser":
		userToGet := getUserData(request.Parameters)
		putTokenInContextIfAny(&ctx, request.Parameters)
		result, err = clientsGRPC.userService.GetUserByID(ctx, userToGet)

	case "AuthUser":
		user := getUserData(request.Parameters)
		result, err = clientsGRPC.userService.AuthUser(ctx, user)

	case "SetRole":
		user := getUserData(request.Parameters)
		putTokenInContextIfAny(&ctx, request.Parameters)
		result, err = clientsGRPC.userService.SetUsersRole(ctx, user)

	case "GetAllProblems":
		putTokenInContextIfAny(&ctx, request.Parameters)
		result, err = clientsGRPC.problemService.GetAllProblems(ctx, &protoProblem.ProblemRequest{})

	case "CreateProblem":
		putTokenInContextIfAny(&ctx, request.Parameters)
		problem := getProblemData(request.Parameters)
		result, err = clientsGRPC.problemService.AddNewProblem(ctx, problem)

	case "SolveProblem":
		putTokenInContextIfAny(&ctx, request.Parameters)
		problemsolution := getProblemSolutionData(request.Parameters)
		result, err = clientsGRPC.problemService.AddProblemSolution(ctx, problemsolution)

	case "ViewSolution":
		putTokenInContextIfAny(&ctx, request.Parameters)
		problem := getProblemData(request.Parameters)
		result, err = clientsGRPC.problemService.GetSolutionByProblem(ctx, problem)

	}

	if err != nil {
		fmt.Fprintf(w, "%s error: %v", request.Method, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func putTokenInContextIfAny(ctx *context.Context, params interface{}) {
	var token string
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return
	}

	tokenParam, ok := mapParam["token"]
	if !ok {
		return
	}

	token, ok = tokenParam.(string)
	if !ok {
		return
	}

	md := metadata.Pairs("authorization", token)
	*ctx = metadata.NewOutgoingContext(*ctx, md)
}

func getUserData(params interface{}) *protoUser.User {
	resUser := &protoUser.User{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resUser
	}

	if id, ok := mapParam["id"]; ok {
		idConv, err := strconv.Atoi(id.(string))
		if err == nil {
			resUser.Id = int64(idConv)
		}
	}
	if name, ok := mapParam["name"]; ok {
		resUser.Name = name.(string)
	}
	if surname, ok := mapParam["surname"]; ok {
		resUser.Surname = surname.(string)
	}
	if email, ok := mapParam["email"]; ok {
		resUser.Email = email.(string)
	}
	if password, ok := mapParam["password"]; ok {
		resUser.Password = password.(string)
	}
	if roleParams, ok := mapParam["role"]; ok {
		resUser.Role = getRoleData(roleParams)
	}

	return resUser
}

func getRoleData(params interface{}) *protoUser.Role {
	resRole := &protoUser.Role{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resRole
	}

	if id, ok := mapParam["id"]; ok {
		idConv, err := strconv.Atoi(id.(string))
		if err == nil {
			resRole.Id = int32(idConv)
		}
	}
	if name, ok := mapParam["name"]; ok {
		resRole.Name = name.(string)
	}
	if isAdmin, ok := mapParam["is_admin"]; ok {
		resRole.IsAdmin = isAdmin.(bool)
	}
	if isCustomer, ok := mapParam["is_customer"]; ok {
		resRole.IsCustomer = isCustomer.(bool)
	}
	if isSupplier, ok := mapParam["is_supplier"]; ok {
		resRole.IsSupplier = isSupplier.(bool)
	}

	return resRole
}

func getProblemData(params interface{}) *protoProblem.Problem {
	resProblem := &protoProblem.Problem{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resProblem
	}

	if id, ok := mapParam["id"]; ok {
		idConv, err := strconv.Atoi(id.(string))
		if err == nil {
			resProblem.Id = int64(idConv)
		}
	}
	if userid, ok := mapParam["user_id"]; ok {
		idConv, err := strconv.Atoi(userid.(string))
		if err == nil {
			resProblem.UserId = int64(idConv)
		}
	}
	if description, ok := mapParam["description"]; ok {
		resProblem.Description = description.(string)
	}
	if solved, ok := mapParam["is_solved"]; ok {
		resProblem.IsSolved = solved.(bool)
	}
	if roleParams, ok := mapParam["type"]; ok {
		resProblem.Type = getTypeData(roleParams)
	}

	return resProblem
}

func getTypeData(params interface{}) *protoProblem.ProblemType {
	resType := &protoProblem.ProblemType{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resType
	}

	if id, ok := mapParam["id"]; ok {
		idConv, err := strconv.Atoi(id.(string))
		if err == nil {
			resType.Id = int32(idConv)
		}
	}
	if name, ok := mapParam["name"]; ok {
		resType.Name = name.(string)
	}

	return resType
}

func getSolutionData(params interface{}) *protoProblem.Solution {
	resSolution := &protoProblem.Solution{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resSolution
	}

	if description, ok := mapParam["description"]; ok {
		resSolution.Description = description.(string)
	}

	return resSolution
}

func getProblemSolutionData(params interface{}) *protoProblem.ProblemSolution {
	resProblemSolution := &protoProblem.ProblemSolution{}
	mapParam, ok := params.(map[string]interface{})
	if !ok {
		return resProblemSolution
	}

	if problem, ok := mapParam["problem"]; ok {
		resProblemSolution.Problem = getProblemData(problem)
	}
	if solution, ok := mapParam["solution"]; ok {
		resProblemSolution.Solution = getSolutionData(solution)
	}

	return resProblemSolution
}
