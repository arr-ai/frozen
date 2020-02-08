package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/arr-ai/frozen"
	"github.com/arr-ai/frozen/slave/proto/slave"
)

type slaveServer struct {
}

func (s *slaveServer) Merge(_ context.Context, maps *slave.Set) (*slave.Map, error) {
	var result frozen.Map
	for i := fromSet(maps).Range(); i.Next(); {
		result = result.Merge(i.Value().(frozen.Map), func(_, a, b interface{}) interface{} {
			return a
		})
	}
	return nil, fmt.Errorf("unfinished")
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
	// determine whether to use TLS
	panic(grpcServer.Serve(skt))
}
