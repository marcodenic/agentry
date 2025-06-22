package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"

	"github.com/marcodenic/agentry/internal/hubgrpc"
)

func main() {
	addr := flag.String("addr", ":9090", "listen address")
	nodes := flag.String("nodes", "", "comma-separated node addresses")
	flag.Parse()

	addrs := []string{}
	if *nodes != "" {
		addrs = strings.Split(*nodes, ",")
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	hubgrpc.Register(srv, hubgrpc.New(addrs))
	log.Printf("hub listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
