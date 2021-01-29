package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
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

func main() {
	server1Addr := flag.String("s1", "localhost:12345", "server address <HOSTNAME>:<PORT>")
	server2Addr := flag.String("s2", "localhost:12345", "server address <HOSTNAME>:<PORT>")
	numRows := flag.Int("n", 10000, "Num DB rows")
	rowLength := flag.Int("r", 32, "Row length in bytes")
	numWorkers := flag.Int("w", 2, "Num workers")
	pirTypeStr := flag.String("t", "punc", fmt.Sprintf("PIR type: [%s]", strings.Join(b.PirTypeStrings(), "|")))
	hintProf := flag.String("hintprof", "", "Profile Server.Hint filename")
	answerProf := flag.String("answerprof", "", "Profile Server.Answer filename")
	updatable := flag.Bool("updatable", true, "Use Updatable PIR")
	latenciesFile := flag.String("latenciesFile", "", "Latencies output filename")

	flag.Parse()

	fmt.Printf("Connecting to %s...", *server1Addr)
	proxyLeft, err := b.NewPirRpcProxy(*server1Addr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	fmt.Printf("Connecting to %s...", *server2Addr)
	proxyRight, err := b.NewPirRpcProxy(*server2Addr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Setting up remote DB...")
	var none int

	config := b.TestConfig{NumRows: *numRows, RowLen: *rowLength, Updatable: *updatable, RandSeed: 678}
	config.PirType, err = b.PirTypeString(*pirTypeStr)
	if err != nil {
		log.Fatalf("Bad PirType: %s", *pirTypeStr)
	}

	for i := 0; i < NumDifferentReads; i++ {
		idx := i % *numRows
		value := make([]byte, *rowLength)
		rand.Read(value)
		config.PresetRows = append(config.PresetRows, b.RowIndexVal{
			Index: idx,
			Key:   rand.Uint32(),
			Value: value})
	}

	client := b.NewPirClientUpdatable(b.RandSource(), [2]b.PirServer{proxyLeft, proxyRight})
	err = proxyLeft.Configure(config, &none)
	if err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}
	err = proxyRight.Configure(config, &none)
	if err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}

	fmt.Printf("[OK]\n")

	fmt.Printf("Obtaining hint (this may take a while)...")
	if len(*hintProf) > 0 {
		err = proxyLeft.StartCpuProfile(0, &none)
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
		err = proxyLeft.StopCpuProfile(0, &profOut)
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
	proxyLeft.ShouldRecord = true
	proxyRight.ShouldRecord = true
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
	proxyLeft.ShouldRecord = false
	proxyRight.ShouldRecord = false
	fmt.Printf("(%d #cached) [OK]\n", len(proxyRight.QueryReqs))

	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(1 * time.Second)
	var totalNumQueries, totalLatency uint64
	latencies := make(chan int, 1000)

	if len(*answerProf) > 0 {
		err = proxyLeft.StartCpuProfile(0, &none)
		if err != nil {
			log.Fatalf("Failed to StartCpuProfile: %s\n", err)
		}
	}

	for i := 0; i < *numWorkers; i++ {
		go func(idx int) {
			for {
				idx := rand.Intn(len(proxyRight.QueryReqs))
				var queryResp b.QueryResp
				start := time.Now()
				err := proxyRight.Answer(proxyRight.QueryReqs[idx], &queryResp)
				elapsed := time.Since(start)
				if len(*latenciesFile) > 0 {
					latencies <- int(elapsed.Microseconds())
				}
				atomic.AddUint64(&totalLatency, uint64(elapsed))
				if err != nil {
					log.Fatalf("Failed to replay query number %d: %s\n", idx, err)
				}
				if !reflect.DeepEqual(proxyRight.QueryResps[idx], queryResp) {
					log.Fatalf("Mismatching response in query number %d", idx)
				}
				counter.Incr(1)
				atomic.AddUint64(&totalNumQueries, 1)
			}
		}(i)
	}

	if len(*latenciesFile) > 0 {
		go func() {
			lOut, _ := os.Create(*latenciesFile)
			for l := range latencies {
				lOut.WriteString(fmt.Sprintf("%d\n", l))
			}
			lOut.Close()
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if len(*answerProf) > 0 {
			var profOut string
			err = proxyLeft.StopCpuProfile(0, &profOut)
			if err != nil {
				log.Fatalf("Failed to StopCpuProfile: %s\n", err)
			}
			err := ioutil.WriteFile(*answerProf, []byte(profOut), 0644)
			if err != nil {
				log.Fatalf("Failed to write server profile to file: %s\n", err)
			}
			fmt.Printf("\nWrote Server.Answer profile file: %s\n", *answerProf)
		}
		if len(*latenciesFile) > 0 {
			close(latencies)
		}
		os.Exit(0)
	}()

	for {
		var avgLatency uint64
		if totalNumQueries > 0 {
			avgLatency = totalLatency / totalNumQueries
		}
		fmt.Printf("\rCurrent rate: %d QPS, average latency: %.02f ms", counter.Rate(), float64(avgLatency)/1000000)
	}
}
