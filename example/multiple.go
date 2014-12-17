package main

import (
	"github.com/cheggaaa/pb"
	"math/rand"
	"sync"
	"time"
)

func main() {
	pool := &pb.Pool{}
	first := pb.New(1000).Prefix("First ")
	second := pb.New(1000).Prefix("Second ")
	third := pb.New(1000).Prefix("Third ")
	pool.Add(first, second, third)
	pool.Start()
	wg := new(sync.WaitGroup)
	for _, bar := range []*pb.ProgressBar{first, second, third} {
		wg.Add(1)
		go func(cb *pb.ProgressBar) {
			for n := 0; n < 1000; n++ {
				cb.Increment()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
			cb.Finish()
			wg.Done()
		}(bar)
	}
	wg.Wait()
}
