package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/marcodenic/agentry/api"
	"github.com/marcodenic/agentry/internal/hubgrpc"
	"github.com/marcodenic/agentry/internal/nodegrpc"
)

func startNode(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	nodegrpc.Register(srv, nodegrpc.New())
	go srv.Serve(lis)
	return lis.Addr().String(), srv.Stop
}

func startHub(t *testing.T, nodes []string) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	hubgrpc.Register(srv, hubgrpc.New(nodes))
	go srv.Serve(lis)
	return lis.Addr().String(), srv.Stop
}

func TestRemoteSpawnTrace(t *testing.T) {
	nodeAddr, stopNode := startNode(t)
	defer stopNode()
	hubAddr, stopHub := startHub(t, []string{nodeAddr})
	defer stopHub()

	conn, err := grpc.Dial(hubAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := api.NewAgentHubClient(conn)

	ctx := context.Background()
	sp, err := client.Spawn(ctx, &api.SpawnRequest{})
	if err != nil {
		t.Fatal(err)
	}

	stream, err := client.Trace(ctx, &api.TraceRequest{AgentId: sp.AgentId})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.SendMessage(ctx, &api.SendMessageRequest{AgentId: sp.AgentId, Input: "hi"}); err != nil {
		t.Fatal(err)
	}

	recvCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	for {
		ev, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if ev.Type != "" {
			break
		}
		if recvCtx.Err() != nil {
			t.Fatal("timeout waiting for trace event")
		}
	}
}
