package tree

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
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
			addrs := []string{}
			for _, glob := range strings.Split(slaveAddrs, ",") {
				expandRanges(glob, func(addr string) {
					addrs = append(addrs, addr)
					conn, err := grpc.Dial(addr, grpc.WithInsecure())
					if err != nil {
						panic(err)
					}
					client := slave.NewSlaveClient(conn)
					if _, err := client.SetSeed(context.TODO(), &slave.SetSeedRequest{A: a, H: h64}); err != nil {
						panic(err)
					}
					slaveClientsCache = append(slaveClientsCache, client)
				})
			}
			log.Printf("FROZEN_SLAVES=%s", strings.Join(addrs, ","))
		}
	})
	return slaveClientsCache
}

var slaveAddrsRangeRE = regexp.MustCompile(`{(\d+)..(\d+)}`)

func expandRanges(glob string, addr func(string)) {
	if m := slaveAddrsRangeRE.FindStringSubmatchIndex(glob); m != nil {
		a, err := strconv.Atoi(glob[m[2]:m[3]])
		if err != nil {
			panic(err)
		}
		b, err := strconv.Atoi(glob[m[4]:m[5]])
		if err != nil {
			panic(err)
		}
		for i := a; i <= b; i++ {
			expandRanges(fmt.Sprintf("%s%0*d%s", glob[:m[0]], m[3]-m[2], i, glob[m[1]:]), addr)
		}
	} else {
		addr(glob)
	}
}
