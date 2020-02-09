package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arr-ai/frozen/internal/tree"
	"github.com/arr-ai/frozen/slave/proto/slave"
	"github.com/arr-ai/hash"
)

type slaveServer struct {
}

var workers = make(chan int)
var blocks = []rune(" ▏▎▍▌▋▊▉█")
var solidBlock = string(blocks[len(blocks)-1:])

type message struct {
	format string
	args   []interface{}
}

var messages = make(chan message)

func infof(format string, args ...interface{}) {
	messages <- message{format: format, args: args}
}

func reportWorkers() {
	n := 0
	for {
		select {
		case i := <-workers:
			n += i
		case m := <-messages:
			fmt.Printf("\r\033[K")
			log.Printf(m.format, m.args...)
		}
		fmt.Printf("\rworkers: %3d %s%c \033[K", n, strings.Repeat(solidBlock, n/8), blocks[n%8])
	}
}

func (s *slaveServer) SetSeed(_ context.Context, req *slave.SetSeedRequest) (*slave.SetSeedResponse, error) {
	h := make([]uintptr, 0, len(req.H))
	for _, u := range req.H {
		h = append(h, uintptr(u))
	}
	if err := hash.SetSeeds(req.A, h); err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}
	infof("Seeds set by remote: %v, %v", base64.RawStdEncoding.EncodeToString(req.A), h)
	return &slave.SetSeedResponse{}, nil
}

func (s *slaveServer) Compute(_ context.Context, req *slave.Work) (*slave.Result, error) {
	workers <- 1
	defer func() { workers <- -1 }()

	resolver := tree.ResolverByName(req.Resolver)

	a, err := tree.FromSlaveTree(req.A)
	if err != nil {
		return nil, err
	}

	b, err := tree.FromSlaveTree(req.B)
	if err != nil {
		return nil, err
	}

	depth := int(req.Depth)

	matches := 0
	var node *tree.Node
	switch req.Op {
	case slave.Work_OP_INTERSECTION:
		a.Intersection(b, resolver, depth, &matches, tree.Mutator, &node)
	case slave.Work_OP_UNION:
		node = a.Union(b, resolver, depth, &matches, tree.Mutator)
	}
	result, err := tree.ToSlaveTree(node)

	if err != nil {
		return nil, err
	}
	return &slave.Result{
		Result:  result,
		Matches: int64(matches),
	}, nil
}

func main() {
	listen := flag.String("listen", "", "[host]:port to listen on")
	flag.Parse()

	if *listen == "" {
		panic("missing --listen")
	}

	skt, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
					return status.Errorf(codes.Unknown, "panic triggered: %v", p)
				}),
			),
		),
	)
	slave.RegisterSlaveServer(grpcServer, &slaveServer{})
	// TODO: TLS

	go reportWorkers()

	log.Printf("Listening on %s", *listen)
	workers <- 0

	panic(grpcServer.Serve(skt))
}
