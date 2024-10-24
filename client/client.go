package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/vikesh-raj/xds-test/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds" // For xds resolver
)

func main() {
	address := flag.String("host", ":51051", ":51051 or xds:///echo")
	name := flag.String("name", "client", "client name")
	count := flag.Int("count", 100, "number of requests to execute")

	flag.Parse()

	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		fmt.Printf("Unable to make new client : %v", err)
		return
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	for i := 0; i < *count; i++ {
		fmt.Printf(">>>>>>>>>>>>>>>>>>>>> %d ", i)
		arg := fmt.Sprintf("%s-%d", *name, i)
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: arg})
		if err != nil {
			fmt.Printf("Error : Unable to connect : %v\n", err)
		} else {
			fmt.Printf("RPC Response(%d): %v\n", i, r.Message)
		}
		time.Sleep(2 * time.Second)
	}
}
