go-proto:
	@go get google.golang.org/grpc google.golang.org/protobuf
	@protoc -I=./proto --go_out=./proto ./proto/*.proto --go-grpc_out=./proto ./proto/*.proto

go-build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/httpserver

docker-up:
	@docker-compose up --build -d

docker-down:
	@docker-compose down