package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/arr-ai/frozen/slave/proto/slave"
)

type slaveServer struct {
}

func (s *slaveServer) Union(_ context.Context, req *slave.UnionRequest) (*slave.Tree, error) {
	// a := fromTree(req.A)
	// b := fromTree(req.B)
	switch req.Op {
	case slave.UnionRequest_OP_UNSPECIFIED:
		return nil, fmt.Errorf("unspecified operation")
	case slave.UnionRequest_OP_CARTESIAN_PRODUCT:
		panic("unfinished")
	default:
		panic(fmt.Errorf("unknown op: %d", req.Op))
	}
}

func main() {
	listen := flag.String("listen", "", "[host]:port to listen on")
	flag.Parse()

	skt, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	slave.RegisterSlaveServer(grpcServer, &slaveServer{})
	// TODO: TLS
	panic(grpcServer.Serve(skt))
}
