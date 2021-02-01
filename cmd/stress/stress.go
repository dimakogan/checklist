package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"github.com/paulbellamy/ratecounter"

	"log"
)

// Number of different records to read to avoid caching effects.
var NumDifferentReads = 100
var numRows int
var updateSize int

//go:generate enumer -type=LoadType
type LoadType int

const (
	Answer LoadType = iota
	Hint
	KeyUpdate
)

var pirType b.PirType

func main() {
	serverAddr := flag.String("s", "localhost:12345", "server address <HOSTNAME>:<PORT>")
	flag.IntVar(&numRows, "n", 10000, "Num DB rows")
	rowLength := flag.Int("r", 32, "Row length in bytes")
	numWorkers := flag.Int("w", 2, "Num workers")
	useTLS := flag.Bool("tls", true, "Should use TLS")
	loadTypeStr := flag.String("l", Answer.String(), "load type: Answer|Hint|KeyUpdate")
	flag.IntVar(&updateSize, "b", 500, "Number of new keys on each key update")
	usePersistent := flag.Bool("persistent", false, "Should use persistent connections")
	pirTypeStr := flag.String("t", "punc", fmt.Sprintf("PIR type: [%s]", strings.Join(b.PirTypeStrings(), "|")))
	clientProf := flag.String("clientprof", "", "Profile Client filename")
	hintProf := flag.String("hintprof", "", "Profile Server.Hint filename")
	answerProf := flag.String("answerprof", "", "Profile Server.Answer filename")
	updatable := flag.Bool("updatable", true, "Use Updatable PIR")

	flag.Parse()

	loadType, err := LoadTypeString(*loadTypeStr)
	if err != nil {
		log.Fatalf("Bad LoadType: %s\n", loadTypeStr)
	}

	fmt.Printf("Connecting to %s...", *serverAddr)
	proxy, err := b.NewPirRpcProxy(*serverAddr, *useTLS, *usePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Setting up remote DB...")
	var none int

	config := b.TestConfig{NumRows: numRows, RowLen: *rowLength, Updatable: *updatable, RandSeed: 678}

	for i := 0; i < NumDifferentReads; i++ {
		idx := i % numRows
		value := make([]byte, *rowLength)
		rand.Read(value)
		config.PresetRows = append(config.PresetRows, b.RowIndexVal{
			Index: idx,
			Key:   rand.Uint32(),
			Value: value})
	}

	pirType, err = b.PirTypeString(*pirTypeStr)
	if err != nil {
		log.Fatalf("Bad PirType: %s", *pirTypeStr)
	}
	client := b.NewPirClientUpdatable(b.RandSource(), pirType, [2]b.PirUpdatableServer{proxy, proxy})
	err = proxy.Configure(config, &none)
	if err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}

	fmt.Printf("[OK]\n")

	var sizes []int
	var probs []float64
	if loadType == Hint {
		sizes = client.LayersMaxSize(numRows)
		probs = make([]float64, len(sizes))
		probs[len(probs)-1] = 1.0
		overflowSize := 2 * sizes[len(sizes)-1]
		for l := len(sizes) - 2; l >= 0; l-- {
			// Number of time this layer gots activated before it overflows
			c := sizes[l] / overflowSize
			probs[l] = probs[l+1] * (float64(c) / float64(c+1))
			overflowSize = (c + 1) * overflowSize
		}
		fmt.Printf("Using layer sizes %v with probabilities %v\n", sizes, probs)
	}

	fmt.Printf("Obtaining hint (this may take a while)...")
	if len(*hintProf) > 0 {
		err = proxy.StartCpuProfile(0, &none)
		if err != nil {
			log.Fatalf("Failed to StartCpuProfile: %s\n", err)
		}
	}
	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}
	if len(*hintProf) > 0 {
		var profOut string
		err = proxy.StopCpuProfile(0, &profOut)
		if err != nil {
			log.Fatalf("Failed to StopCpuProfile: %s\n", err)
		}
		err := ioutil.WriteFile(*hintProf, []byte(profOut), 0644)
		if err != nil {
			log.Fatalf("Failed to write server profile to file: %s\n", err)
		}
		log.Printf("Wrote Server.Hint profile file: %s\n", *hintProf)
	}

	fmt.Printf("[OK]\n")

	fmt.Printf("Caching responses...")
	proxy.ShouldRecord = true
	for i := 0; i < NumDifferentReads; i++ {
		idx := rand.Intn(NumDifferentReads)
		readVal, err := client.Read(config.PresetRows[idx].Key)
		if err != nil {
			log.Fatalf("Failed to read index %d: %s", i, err)
		}
		if !reflect.DeepEqual(config.PresetRows[idx].Value, readVal) {
			log.Fatalf("Mismatching row value at index %d", idx)
		}
	}
	proxy.ShouldRecord = false
	fmt.Printf("(%d #cached) [OK]\n", len(proxy.QueryReqs))

	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(10 * time.Second)
	startTime := time.Now()
	var totalNumQueries, totalLatency uint64

	if len(*answerProf) > 0 {
		err = proxy.StartCpuProfile(0, &none)
		if err != nil {
			log.Fatalf("Failed to StartCpuProfile: %s\n", err)
		}
	}

	for i := 0; i < *numWorkers; i++ {
		go func(idx int) {
			for {
				start := time.Now()
				switch loadType {
				case Answer:
					replayQuery(proxy)
				case Hint:
					replayHint(proxy, sizes, probs)
				case KeyUpdate:
					replayKeyUpdate(proxy)
				}

				elapsed := time.Since(start)
				atomic.AddUint64(&totalLatency, uint64(elapsed))

				counter.Incr(1)
				atomic.AddUint64(&totalNumQueries, 1)
			}
		}(i)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		var f *os.File
		if len(*clientProf) > 0 {
			f, err = os.Create(*clientProf)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
		}

		<-c

		if f != nil {
			pprof.StopCPUProfile()
			f.Close()
		}

		if len(*answerProf) > 0 {
			var profOut string
			err = proxy.StopCpuProfile(0, &profOut)
			if err != nil {
				log.Fatalf("Failed to StopCpuProfile: %s\n", err)
			}
			err := ioutil.WriteFile(*answerProf, []byte(profOut), 0644)
			if err != nil {
				log.Fatalf("Failed to write server profile to file: %s\n", err)
			}
			fmt.Printf("\nWrote Server.Answer profile file: %s\n", *answerProf)
		}
		os.Exit(0)
	}()

	for {
		var avgLatency uint64
		if totalNumQueries > 0 {
			avgLatency = totalLatency / totalNumQueries
		}
		time.Sleep(time.Second)
		fmt.Printf("\rCurrent rate: %.02f QPS, overall rate: %.02f, average latency: %.02f ms",
			float64(counter.Rate())/10,
			float64(totalNumQueries)/time.Since(startTime).Seconds(),
			float64(avgLatency)/1000000)
	}
}

func replayQuery(proxy *b.PirRpcProxy) error {
	idx := rand.Intn(len(proxy.QueryReqs))
	var queryResp b.QueryResp
	err := proxy.Answer(proxy.QueryReqs[idx], &queryResp)
	if err != nil {
		return fmt.Errorf("Failed to replay query number %d to server: %s", idx, err)
	}
	if !reflect.DeepEqual(proxy.QueryResps[idx], queryResp) {
		return fmt.Errorf("Mismatching  response in query number %d", idx)
	}
	return nil
}

func replayKeyUpdate(proxy *b.PirRpcProxy) error {
	keyReq := b.KeyUpdatesReq{
		DefragTimestamp: math.MaxInt32,
		NextTimestamp:   int32(numRows - updateSize),
	}
	//fmt.Printf("Using size: %d\n", layerSize)
	var keyResp b.KeyUpdatesResp
	err := proxy.KeyUpdates(keyReq, &keyResp)
	if err != nil {
		return fmt.Errorf("Failed to replay key update request %v, %s", keyReq, err)
	}
	if len(keyResp.Keys) != updateSize {
		return fmt.Errorf("Invalid size of key update, expected: %d, got: %d", numRows, len(keyResp.Keys))
	}
	return nil
}

func randSize(sizes []int, probs []float64) int {
	p := rand.Float64()
	bucket := 0
	for p > probs[bucket] {
		bucket++
	}
	return sizes[bucket]
}

func replayHint(proxy *b.PirRpcProxy, sizes []int, probs []float64) error {
	layerSize := randSize(sizes, probs)
	firstRow := rand.Intn(numRows - layerSize + 1)
	hintReq := b.HintReq{
		RandSeed:        42,
		DefragTimestamp: math.MaxInt32,
		Layers:          []b.HintLayer{{FirstRow: firstRow, NumRows: layerSize, PirType: pirType}},
	}
	//fmt.Printf("Using size: %d\n", layerSize)
	var hintResp b.HintResp
	err := proxy.Hint(hintReq, &hintResp)
	if err != nil {
		return fmt.Errorf("Failed to replay hint request %v, %s", hintReq, err)
	}
	if len(hintResp.BatchResps) < 1 {
		return fmt.Errorf("Failed to replay hint request, 0 subresponses: %v", hintReq)
	}
	if hintResp.BatchResps[0].NumRows != numRows {
		return fmt.Errorf("Failed to replay hint request %v , mismatching hint num rows, expected: %d, got: %d", hintReq, numRows, hintResp.BatchResps[0].NumRows)
	}
	return nil
}
