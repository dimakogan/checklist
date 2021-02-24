package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	sleepMsec       int
	totalNumQueries uint64
	totalNumErrors  uint64
	numWorkers      int
	wg              sync.WaitGroup

	reqs, errors                            *ratecounter.RateCounter
	latency                                 *ratecounter.AvgRateCounter
	reqsMet, errMet, latencyMet, workersMet metric.Metric
}

func initContext() *workerCtx {
	return &workerCtx{
		startTime: time.Now(),
		// We're recording marks-per-10second
		reqs:       ratecounter.NewRateCounter(time.Second),
		errors:     ratecounter.NewRateCounter(time.Second),
		latency:    ratecounter.NewAvgRateCounter(time.Second),
		reqsMet:    metric.NewCounter("1m1s", "5m10s"),
		errMet:     metric.NewCounter("1m1s", "5m10s"),
		latencyMet: metric.NewGauge("1m1s", "5m10s"),
		workersMet: metric.NewCounter("1m1s", "5m10s"),
	}
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
		if err != nil {
			ctx.errors.Incr(1)
			ctx.errMet.Add(1)
			atomic.AddUint64(&ctx.totalNumErrors, 1)
			continue
		}

		elapsed := time.Since(start).Milliseconds()
		ctx.reqs.Incr(1)
		ctx.latency.Incr(elapsed)
		ctx.reqsMet.Add(1)
		ctx.latencyMet.Add(float64(elapsed))
		atomic.AddUint64(&ctx.totalNumQueries, 1)
		time.Sleep(time.Duration(ctx.sleepMsec) * time.Millisecond)
	}
}

func runWorkers(config *Configurator, stresser stresser, ctx *workerCtx, numWorkers []int, incInterval int) {
	for _, w := range numWorkers {
		toAdd := w - ctx.numWorkers
		ctx.wg.Add(toAdd)
		ctx.workersMet.Add(float64(toAdd))
		ctx.numWorkers = w

		for i := 0; i < toAdd; i++ {
			go workerFunc(config, stresser, ctx)
		}
		if incInterval == 0 {
			return
		}
		time.Sleep(time.Duration(incInterval) * time.Second)
		if ctx.inShutdown {
			break
		}
	}
}

func liveMonitor(ctx *workerCtx, outFile string) {
	expvar.Publish("requests", ctx.reqsMet)
	expvar.Publish("latency", ctx.latencyMet)
	expvar.Publish("workers", ctx.workersMet)
	expvar.Publish("errors", ctx.errMet)
	http.Handle("/debug/metrics", metric.Handler(metric.Exposed))
	go http.ListenAndServe(":8080", nil)

	var f *os.File
	var err error
	if outFile != "" {
		f, err = os.OpenFile(outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to create output file: %s", err)
		}
		defer f.Close()
	}

	fmt.Fprintf(f, "Seconds,Workers,Queries,Latency,Errors\n")

	for {
		if ctx.inShutdown {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf("\rWorkers: %d, Current rate: %d QPS, overall rate (over %s): %.02f, average latency: %.02f ms, errors: %d          ",
			ctx.numWorkers,
			ctx.reqs.Rate(),
			time.Since(ctx.startTime).String(),
			float64(ctx.totalNumQueries)/time.Since(ctx.startTime).Seconds(),
			float64(ctx.latency.Rate()),
			ctx.errors.Rate())
		if f != nil {
			fmt.Fprintf(f, "%d,%d,%d,%.02f,%d\n", time.Now().Unix(), ctx.numWorkers, ctx.totalNumQueries, ctx.latency.Rate(), ctx.totalNumErrors)
		}
	}
	ctx.wg.Wait()
}

func main() {
	config := NewConfig().WithClientFlags()
	numWorkers := config.FlagSet.String("w", "2", "Num workers (sequence)")
	incInterval := config.FlagSet.Int("i", 0, "Interval to increment num workers")
	sleepMsec := config.FlagSet.Int("s", 500, "milliseconds to sleep between requests")
	loadTypeStr := config.FlagSet.String("l", Answer.String(), "load type: Answer|Hint|KeyUpdate")
	outFile := config.FlagSet.String("o", "", "output filename for stats")
	// config.FlagSet.String("recordTo", "", "File to store recorded requests at.")
	config.Parse()

	numWorkersSeq := []int{}
	wstrs := strings.Split(*numWorkers, ",")
	for _, wstr := range wstrs {
		if w, err := strconv.Atoi(wstr); err != nil {
			log.Fatalf("Invalid num workers value: %s", wstr)
		} else {
			numWorkersSeq = append(numWorkersSeq, w)
		}
	}

	loadType, err := LoadTypeString(*loadTypeStr)
	if err != nil {
		log.Fatalf("Bad LoadType: %s\n", *loadTypeStr)
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
	ctx := initContext()
	ctx.sleepMsec = *sleepMsec

	go runWorkers(config, stresser, ctx, numWorkersSeq, *incInterval)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		prof := NewProfiler(config.CpuProfile)
		defer prof.Close()
		<-c
		ctx.inShutdown = true
	}()

	liveMonitor(ctx, *outFile)
}
