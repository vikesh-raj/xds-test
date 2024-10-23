package main

import (
	"flag"
	"fmt"

	"github.com/vikesh-raj/xds-test/echo"

	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds" // For xds resolver
)

func main() {
	address := flag.String("host", ":51051", ":51051 or xds:///localhost:51050")
	name := flag.String("name", "client", "client name")

	flag.Parse()

	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	for i := 0; i < 30; i++ {
		arg := fmt.Sprintf("%s-%d", *name, i)
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: arg})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("RPC Response: %v %v", i, r)
		time.Sleep(2 * time.Second)
	}
}
