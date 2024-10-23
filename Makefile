.PHONY: proto build

proto:
	protoc -I=echo --go_out=echo --go_opt=paths=source_relative --go-grpc_out=echo --go-grpc_opt=paths=source_relative echo/echo.proto

build:
	cd echo && go build
	cd client && go build
	cd server && go build

install_proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

tidy:
	cd echo && go mod tidy
	cd client && go mod tidy
	cd server && go mod tidy