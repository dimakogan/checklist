package main

import (
	"expvar"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	. "checklist/driver"

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
	User
)

type loadGen interface {
	request(proxy *RpcProxy) error
	reqRate() int
	debugStr() string
}

type testConfig struct {
	Config

	loadType      LoadType
	numWorkersSeq []int
	incInterval   int
	sleepMsec     int
	outFile       string
}

type stressTest struct {
	testConfig

	// State
	load          loadGen
	inShutdown    bool
	addingWorkers bool
	curNumWorkers int
	wg            sync.WaitGroup

	// Monitoring
	startTime                               time.Time
	reqs, errors                            *ratecounter.RateCounter
	latency                                 *ratecounter.AvgRateCounter
	reqsMet, errMet, latencyMet, workersMet metric.Metric
	totalNumQueries                         uint64
	totalNumErrors                          uint64
}

func parseFlags(config *testConfig) {
	config.AddPirFlags().AddClientFlags()
	config.FlagSet.StringVar(&config.TraceFile, "trace", "", "input trace file")
	config.FlagSet.IntVar(&config.incInterval, "i", 0, "Interval to increment num workers")
	config.FlagSet.IntVar(&config.sleepMsec, "s", 500, "milliseconds to sleep between requests")
	config.FlagSet.StringVar(&config.outFile, "o", "", "output filename for stats")

	numWorkersStr := config.FlagSet.String("w", "2", "Num workers (sequence)")
	loadTypeStr := config.FlagSet.String("l", Answer.String(), "load type: Answer|Hint|KeyUpdate")

	config.Parse()

	wstrs := strings.Split(*numWorkersStr, ",")
	for _, wstr := range wstrs {
		if w, err := strconv.Atoi(wstr); err != nil {
			log.Fatalf("Invalid num workers value: %s", wstr)
		} else {
			config.numWorkersSeq = append(config.numWorkersSeq, w)
		}
	}
	var err error
	config.loadType, err = LoadTypeString(*loadTypeStr)
	if err != nil {
		log.Fatalf("Bad LoadType: %s\n", *loadTypeStr)
	}
}

func (t *stressTest) initMonitoring() {
	t.startTime = time.Now()
	// We're recording marks-per-10second
	t.reqs = ratecounter.NewRateCounter(time.Second)
	t.errors = ratecounter.NewRateCounter(time.Second)
	t.latency = ratecounter.NewAvgRateCounter(time.Second)
	t.reqsMet = metric.NewCounter("1m1s", "5m10s")
	t.errMet = metric.NewCounter("1m1s", "5m10s")
	t.latencyMet = metric.NewGauge("1m1s", "5m10s")
	t.workersMet = metric.NewCounter("1m1s", "5m10s")
}

func (t *stressTest) onCompletedReq(latency int64) {
	t.reqs.Incr(1)
	t.latency.Incr(latency)
	t.reqsMet.Add(1)
	t.latencyMet.Add(float64(latency))
	atomic.AddUint64(&t.totalNumQueries, 1)
}

func (t *stressTest) onError() {
	t.errors.Incr(1)
	t.errMet.Add(1)
	atomic.AddUint64(&t.totalNumErrors, 1)
}

func (t *stressTest) addPresetRows() {
	for i := 0; i < NumDifferentReads; i++ {
		idx := i % t.NumRows
		value := make([]byte, t.RowLen)
		rand.Read(value)
		t.PresetRows = append(t.PresetRows, RowIndexVal{
			Index: idx,
			Key:   rand.Uint32(),
			Value: value})
	}
}

func (t *stressTest) workerFunc() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Worker panic: %+v\n%s", r, string(debug.Stack()))
		}
	}()
	proxy, err := NewRpcProxy(t.ServerAddr, t.UseTLS, t.UsePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	for {
		if t.inShutdown {
			t.wg.Done()
			return
		}
		var err error
		start := time.Now()
		err = t.load.request(proxy)
		if err != nil {
			t.onError()
			log.Fatalf("%s", err)
			continue
		}

		t.onCompletedReq(time.Since(start).Milliseconds())
		time.Sleep(time.Duration(t.sleepMsec) * time.Millisecond)
	}
}

func (t *stressTest) runWorkers() {
	for _, w := range t.numWorkersSeq {
		t.addingWorkers = true
		time.Sleep(time.Second)
		toAdd := w - t.curNumWorkers
		t.wg.Add(toAdd)
		t.workersMet.Add(float64(toAdd))
		t.curNumWorkers = w

		for i := 0; i < toAdd; i++ {
			go t.workerFunc()
		}
		time.Sleep(15 * time.Second)
		t.addingWorkers = false
		if t.incInterval == 0 {
			return
		}
		time.Sleep(time.Duration(t.incInterval) * time.Second)
		if t.inShutdown {
			break
		}
	}
}

func (t *stressTest) liveMonitor() {
	expvar.Publish("requests", t.reqsMet)
	expvar.Publish("latency", t.latencyMet)
	expvar.Publish("workers", t.workersMet)
	expvar.Publish("errors", t.errMet)
	http.Handle("/debug/metrics", metric.Handler(metric.Exposed))
	go http.ListenAndServe(":8080", nil)

	var f *os.File
	var err error
	if t.outFile != "" {
		f, err = os.OpenFile(t.outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to create output file: %s", err)
		}
		defer f.Close()
	}

	fmt.Fprintf(f, "Seconds,Workers,Queries,Latency,Errors,Users,Debug\n")

	for {
		if t.inShutdown {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf("\rWorkers: %d, Current rate: %d Req/Sec, overall rate (over %s): %.02f, average latency: %.02f ms, errors: %d          ",
			t.curNumWorkers,
			t.reqs.Rate(),
			time.Since(t.startTime).String(),
			float64(t.totalNumQueries)/time.Since(t.startTime).Seconds(),
			float64(t.latency.Rate()),
			t.errors.Rate())

		if t.addingWorkers {
			continue
		}
		if f != nil {
			fmt.Fprintf(f, "%d,%d,%d,%.02f,%d,%d,%s\n",
				time.Now().Unix(),
				t.curNumWorkers,
				t.totalNumQueries,
				t.latency.Rate(),
				t.totalNumErrors,
				t.totalNumQueries*uint64(t.load.reqRate()),
				t.load.debugStr())
		}
	}
	t.wg.Wait()
}

func (t *stressTest) notifyOnSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		prof := NewProfiler(t.CpuProfile)
		defer prof.Close()
		<-c
		t.inShutdown = true
	}()
}

func main() {
	test := stressTest{}
	parseFlags(&test.testConfig)
	test.DataRandSeed = 678
	test.addingWorkers = true

	switch test.loadType {
	case Hint:
		test.load = initHintLoadGen(&test.Config)
	case Answer:
		test.load = initAnswerLoadGen(&test.Config)
	case KeyUpdate:
		test.load = initKeyUpdateLoadGen(&test.Config)
	case User:
		test.load = initUserLoadGen(&test.Config, LoadTraceFile(test.Config.TraceFile))
	}

	test.initMonitoring()
	go test.runWorkers()
	test.notifyOnSignal()
	test.liveMonitor()
}
