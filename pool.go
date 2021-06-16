package frozen

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var usePools bool

type poolStats struct {
	m   sync.Mutex
	d   map[string]*poolStat
	die chan struct{}
}

var thePoolStats = newPoolStats()

func newPoolStats() *poolStats {
	s := &poolStats{}

	s.d = map[string]*poolStat{}
	s.die = make(chan struct{})

	usePools = os.Getenv("FROZEN_NO_POOLS") != "1"
	logInterval := os.Getenv("FROZEN_POOL_LOG_INTERVAL")
	if poolLogInterval, err := time.ParseDuration(logInterval); err == nil {
		log.Printf("Logging pool stats every %v", poolLogInterval)
		go func() {
			ticker := time.NewTicker(poolLogInterval)
			for {
				select {
				case _, ok := <-s.die:
					if !ok {
						return
					}
				case <-ticker.C:
					s.Report()
				}
			}
		}()
	}
	return s
}

func (s *poolStats) Close() {
	close(s.die)
}

func (s *poolStats) Get(name string) {
	atomic.AddUint64(&s.stat(name).gets, 1)
}

func (s *poolStats) New(name string) {
	atomic.AddUint64(&s.stat(name).news, 1)
}

func (s *poolStats) Put(name string) {
	atomic.AddUint64(&s.stat(name).puts, 1)
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
