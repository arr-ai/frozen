package pool

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// UsePools triggers use of pools to allocate transient node structures.
var UsePools bool

type poolStats struct {
	m   sync.Mutex
	d   map[string]*poolStat
	die chan struct{}
}

// ThePoolStats is the global pool statistics collector.
var ThePoolStats = newPoolStats()

func newPoolStats() *poolStats {
	s := &poolStats{}
	s.d = map[string]*poolStat{}

	UsePools = os.Getenv("FROZEN_NO_POOLS") != "1"
	logInterval := os.Getenv("FROZEN_POOL_LOG_INTERVAL")
	if poolLogInterval, err := time.ParseDuration(logInterval); err == nil {
		log.Printf("Logging pool stats every %v", poolLogInterval)
		s.Start(poolLogInterval)
	}
	return s
}

func (s *poolStats) Running() bool {
	return s.die != nil
}

func (s *poolStats) Start(poolLogInterval time.Duration) {
	if s.die == nil {
		s.die = make(chan struct{})
		die := s.die
		go func() {
			ticker := time.NewTicker(poolLogInterval)
			for {
				select {
				case _, ok := <-die:
					if !ok {
						return
					}
				case <-ticker.C:
					s.Report()
				}
			}
		}()
	}
}

func (s *poolStats) Stop() {
	close(s.die)
	s.die = nil
}

func (s *poolStats) Get(name string) {
	if s.Running() {
		atomic.AddUint64(&s.stat(name).gets, 1)
	}
}

func (s *poolStats) New(name string) {
	if s.Running() {
		atomic.AddUint64(&s.stat(name).news, 1)
	}
}

func (s *poolStats) Put(name string) {
	if s.Running() {
		atomic.AddUint64(&s.stat(name).puts, 1)
	}
}

func (s *poolStats) Report() {
	s.m.Lock()
	defer s.m.Unlock()
	var sb strings.Builder
	for name, stat := range s.d {
		hits := stat.gets - stat.news
		misses := stat.news
		wild := stat.gets - stat.puts
		fmt.Fprintf(&sb, "%s:+%d↓%d↑%d  ", name, hits, misses, wild)
	}
	log.Printf("Pool stats: %v", sb.String())
}

func (s *poolStats) stat(name string) *poolStat {
	s.m.Lock()
	defer s.m.Unlock()
	stat := s.d[name]
	if stat == nil {
		stat = &poolStat{}
		s.d[name] = stat
	}
	return stat
}

type poolStat struct {
	gets uint64
	news uint64
	puts uint64
}
