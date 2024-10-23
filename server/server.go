package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/vikesh-raj/xds-test/echo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type server struct {
	echo.UnsafeEchoServerServer
	name string
}

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {
	log.Println("Got rpc: --> ", in.Name)
	return &echo.EchoReply{Message: "Hello " + in.Name + " from " + s.name}, nil
}

func (s *server) SayHelloStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloStreamServer) error {
	log.Println("Got stream:  -->  ")
	stream.Send(&echo.EchoReply{Message: "Response 1: Hello " + in.Name})
	stream.Send(&echo.EchoReply{Message: "Response 2: Hello " + in.Name})
	return nil
}

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func main() {
	port := flag.Int("port", 51051, "port")
	name := flag.String("name", "echo-server", "server name")
	flag.Parse()

	grpc_port := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", grpc_port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	echo.RegisterEchoServerServer(s, &server{
		name: *name,
	})

	healthpb.RegisterHealthServer(s, &healthServer{})
	log.Println("Starting grpc echo Server on port : " + grpc_port)
	s.Serve(lis)
}
