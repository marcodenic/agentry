package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/marcodenic/agentry/internal/nodegrpc"
)

func main() {
	addr := flag.String("addr", ":9091", "listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	nodegrpc.Register(srv, nodegrpc.New())
	log.Printf("agent node listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
