package main

import (
	"fmt"
	"time"
)

var Watch Stopwatch

type Stopwatch struct {
	Buckets      map[string]int64
	BucketStarts map[string]int64
}

func init() {
	Watch = Stopwatch{}
	Watch.Buckets = make(map[string]int64)
	Watch.BucketStarts = make(map[string]int64)
}

func (s *Stopwatch) Start(b string) {
	s.BucketStarts[b] = time.Now().UnixNano()
	_, ok := s.Buckets[b]
	if !ok {
		s.Buckets[b] = 0
	}
}

func (s *Stopwatch) Stop(b string) {
	end := time.Now().UnixNano()
	start, ok := s.BucketStarts[b]
	if !ok {
		return
	}
	s.Buckets[b] += end - start
	delete(s.BucketStarts, b)
}

func (s *Stopwatch) Results() string {
	out := ""
	for k, v := range s.Buckets {
		out += fmt.Sprintf("%s: %.4f\n", k, float64(v)/1000000000.0)
	}
	return out
}
