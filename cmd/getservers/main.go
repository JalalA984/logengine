package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	api "github.com/jalala984/logengine/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:18000", "service address")
	flag.Parse()

	// Use NewClient for modern gRPC usage (replacing DialContext)
	// and use insecure credentials explicitly.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := api.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		log.Fatalf("could not get servers: %v", err)
	}

	fmt.Printf("Successfully queried cluster at %s\n", *addr)
	fmt.Printf("Cluster size: %d nodes\n", len(res.Servers))
	fmt.Println("Nodes found:")
	for _, server := range res.Servers {
		status := "Follower"
		if server.IsLeader {
			status = "LEADER"
		}
		fmt.Printf("\t- ID: %-12s Addr: %-20s Status: %s\n", 
			server.Id, 
			server.RpcAddr, 
			status,
		)
	}
}