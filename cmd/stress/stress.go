package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	. "github.com/dimakogan/boosted-pir"

	"github.com/paulbellamy/ratecounter"

	"log"
)

// Number of different records to read to avoid caching effects.
var NumDifferentReads = 100

//go:generate enumer -type=LoadType
type LoadType int

const (
	Answer LoadType = iota
	Hint
	KeyUpdate
)

var pirType PirType

type stresser interface {
	request(proxy *PirRpcProxy) error
}

func main() {
	config := NewConfig().WithClientFlags()
	numWorkers := config.FlagSet.Int("w", 2, "Num workers")
	loadTypeStr := config.FlagSet.String("l", Answer.String(), "load type: Answer|Hint|KeyUpdate")
	// config.FlagSet.String("recordTo", "", "File to store recorded requests at.")
	config.Parse()

	loadType, err := LoadTypeString(*loadTypeStr)
	if err != nil {
		log.Fatalf("Bad LoadType: %s\n", loadTypeStr)
	}

	fmt.Printf("Connecting to %s (TLS: %t)...", config.ServerAddr, config.UseTLS)
	proxy, err := NewPirRpcProxy(config.ServerAddr, config.UseTLS, config.UsePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Setting up remote DB...")
	config.RandSeed = 678
	if err := proxy.Configure(config.TestConfig, nil); err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}
	fmt.Printf("[OK]\n")
	var stresser stresser
	switch loadType {
	case Hint:
		stresser = initHintStresser(config, proxy)
	case Answer:
		stresser = initAnswerStresser(config, proxy)
	case KeyUpdate:
		stresser = initKeyUpdateStresser(config, proxy)
	}

	proxy.Close()
	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(10 * time.Second)
	startTime := time.Now()
	var totalNumQueries, totalLatency uint64

	inShutdown := false
	var wg sync.WaitGroup
	wg.Add(*numWorkers)

	for i := 0; i < *numWorkers; i++ {
		go func(idx int) {
			defer func() {
				if r := recover(); r != nil {
					log.Fatalf("Worker panic: %v", r)
				}
			}()
			proxy, err := NewPirRpcProxy(config.ServerAddr, config.UseTLS, config.UsePersistent)
			if err != nil {
				log.Fatal("Connection error: ", err)
			}
			for {
				if inShutdown {
					wg.Done()
					return
				}
				var err error
				start := time.Now()
				err = stresser.request(proxy)
				elapsed := time.Since(start)
				atomic.AddUint64(&totalLatency, uint64(elapsed))

				if err != nil {
					log.Fatalf("Failed to replay query: %v", err)
				}

				counter.Incr(1)
				atomic.AddUint64(&totalNumQueries, 1)
			}
		}(i)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		prof := NewProfiler(config.CpuProfile)
		defer prof.Close()
		<-c
		inShutdown = true
		wg.Wait()
		os.Exit(0)
	}()

	for {
		var avgLatency uint64
		if totalNumQueries > 0 {
			avgLatency = totalLatency / totalNumQueries
		}
		time.Sleep(time.Second)
		fmt.Printf("\rCurrent rate: %.02f QPS, overall rate (over %s): %.02f, average latency: %.02f ms          ",
			float64(counter.Rate())/10,
			time.Since(startTime).String(),
			float64(totalNumQueries)/time.Since(startTime).Seconds(),
			float64(avgLatency)/1000000)
	}
}
