package tree

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/arr-ai/frozen/slave/proto/slave"
	"github.com/arr-ai/hash"
	"google.golang.org/grpc"
)

var slaveAddrs = os.Getenv("FROZEN_SLAVES")

// var slaveAddrs = "localhost:8900"

// var slaveAddrs = ""

var slavesOnce sync.Once
var slaveClientsCache []slave.SlaveClient

func slaveClients() []slave.SlaveClient {
	slavesOnce.Do(func() {
		if slaveAddrs != "" {
			a, h := hash.GetSeeds()
			h64 := make([]uint64, 0, len(h))
			for _, u := range h {
				h64 = append(h64, uint64(u))
			}
			for _, addr := range strings.Split(slaveAddrs, ",") {
				conn, err := grpc.Dial(addr, grpc.WithInsecure())
				if err != nil {
					panic(err)
				}
				client := slave.NewSlaveClient(conn)
				if _, err := client.SetSeed(context.TODO(), &slave.SetSeedRequest{A: a, H: h64}); err != nil {
					panic(err)
				}
				slaveClientsCache = append(slaveClientsCache, client)
			}
		}
	})
	return slaveClientsCache
}
