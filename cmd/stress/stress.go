package main

import (
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	. "github.com/dimakogan/boosted-pir"

	"github.com/paulbellamy/ratecounter"
	"github.com/zserge/metric"
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

type workerCtx struct {
	startTime       time.Time
	inShutdown      bool
	totalNumQueries uint64
	numWorkers      int
	wg              sync.WaitGroup

	log string

	reqs                            *ratecounter.RateCounter
	latency                         *ratecounter.AvgRateCounter
	reqsMet, latencyMet, workersMet metric.Metric
}

func workerFunc(config *Configurator, stresser stresser, ctx *workerCtx) {
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
		if ctx.inShutdown {
			ctx.wg.Done()
			return
		}
		var err error
		start := time.Now()
		err = stresser.request(proxy)
		elapsed := time.Since(start).Milliseconds()
		ctx.reqs.Incr(1)
		ctx.latency.Incr(elapsed)
		ctx.reqsMet.Add(1)
		ctx.latencyMet.Add(float64(elapsed))
		atomic.AddUint64(&ctx.totalNumQueries, 1)

		if err != nil {
			log.Fatalf("Failed to replay query: %v", err)
		}

	}
}

func liveMonitor(ctx *workerCtx) {
	expvar.Publish("requests", ctx.reqsMet)
	expvar.Publish("latency", ctx.latencyMet)
	expvar.Publish("workers", ctx.workersMet)
	http.Handle("/debug/metrics", metric.Handler(metric.Exposed))
	go http.ListenAndServe(":8080", nil)

	for {
		if ctx.inShutdown {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf("\rWorkers: %d, Current rate: %.02f QPS, overall rate (over %s): %.02f, average latency: %.02f ms          ",
			ctx.numWorkers,
			float64(ctx.reqs.Rate()),
			time.Since(ctx.startTime).String(),
			float64(ctx.totalNumQueries)/time.Since(ctx.startTime).Seconds(),
			float64(ctx.latency.Rate()))
		ctx.log += fmt.Sprintf("%d,%d,%d,%.02f\n", int(time.Since(ctx.startTime).Seconds()), ctx.numWorkers, ctx.totalNumQueries, ctx.latency.Rate())
	}
}

func main() {
	config := NewConfig().WithClientFlags()
	numWorkers := config.FlagSet.Int("w", 2, "Num workers")
	loadTypeStr := config.FlagSet.String("l", Answer.String(), "load type: Answer|Hint|KeyUpdate")
	outFile := config.FlagSet.String("o", "", "output filename for stats")
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
	ctx := workerCtx{
		startTime: time.Now(),
		// We're recording marks-per-10second
		reqs:       ratecounter.NewRateCounter(time.Second),
		latency:    ratecounter.NewAvgRateCounter(time.Second),
		reqsMet:    metric.NewCounter("1m1s", "5m10s"),
		latencyMet: metric.NewGauge("1m1s", "5m10s"),
		workersMet: metric.NewCounter("1m1s", "5m10s"),
	}

	go liveMonitor(&ctx)

	ctx.wg.Add(*numWorkers)
	ctx.numWorkers = *numWorkers
	ctx.workersMet.Add(float64(*numWorkers))

	for i := 0; i < *numWorkers; i++ {
		go workerFunc(config, stresser, &ctx)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		prof := NewProfiler(config.CpuProfile)
		defer prof.Close()
		<-c
		ctx.inShutdown = true
		ctx.wg.Wait()

		if *outFile != "" {
			ioutil.WriteFile(*outFile, []byte("Seconds,Workers,Queries,Latency\n"+ctx.log), 0644)
		}
		os.Exit(0)
	}()

	for {
		time.Sleep(15 * time.Second)
		if *outFile != "" {
			ioutil.WriteFile(*outFile, []byte("Seconds,Workers,Queries,Latency\n"+ctx.log), 0644)
		}
		addWorkers := ctx.numWorkers / 2
		ctx.wg.Add(addWorkers)
		ctx.workersMet.Add(float64(addWorkers))
		for i := 0; i < addWorkers; i++ {
			go workerFunc(config, stresser, &ctx)
		}
		ctx.numWorkers += addWorkers
	}
}
