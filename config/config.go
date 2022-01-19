package configs

import (
	"os"
)

var CERT_PATH = getStringParameter("CERT_PATH", "./microservice/certificates/")
var PROBLEMS_GRPC_PORT = getStringParameter("PROBLEMS_GRPC_PORT", "4444")
var PROBLEMS_SERVICE = getStringParameter("PROBLEMS_SERVICE", "localhost")
var PROBLEMS_CERT_NAME = getStringParameter("PROBLEMS_CERT_NAME", "probllocal.crt")
var PROBLEMS_CERTIFICATE = CERT_PATH + PROBLEMS_CERT_NAME

var USERS_GRPC_PORT = getStringParameter("USERS_GRPC_PORT", "5555")
var USER_SERVICE = getStringParameter("USER_SERVICE", "localhost")

func getStringParameter(paramName, defaultValue string) string {
	result, ok := os.LookupEnv(paramName)
	if !ok {
		result = defaultValue
	}
	return result
}