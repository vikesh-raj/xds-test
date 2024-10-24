

build:
    cd echo && go build
    cd client && go build
    cd server && go build
    cd xds-server && go build

proto:
    protoc -I=echo --go_out=echo --go_opt=paths=source_relative --go-grpc_out=echo --go-grpc_opt=paths=source_relative echo/echo.proto

install_proto:
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.1
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

tidy:
    cd echo && go mod tidy
    cd client && go mod tidy
    cd server && go mod tidy
    cd xds-server && go mod tidy

build-client:
    cd client && go build

build-server:
    cd server && go build

build-xds-server:
    cd xds-server && go build

run-server-1:
    ./server/server -port 50051 -name server-1

run-server-2:
    ./server/server -port 50052 -name server-2

run-server-3:
    ./server/server -port 50053 -name server-3

run-xds-server:
    ./xds-server/xds-server -port 50050 -config config.yaml

run-client:
    ./client/client

run-client-with-xds-verbose:
    GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=info GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json ./client/client -host xds:///echo

run-client-with-xds:
    GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json ./client/client -host xds:///echo
