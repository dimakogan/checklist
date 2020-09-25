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
	"syscall"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"github.com/paulbellamy/ratecounter"

	"log"
	"net/rpc"
)

// Number of different records to read to avoid caching effects.
var NumDifferentReads = 100

func main() {
	serverAddr := flag.String("s", "localhost:12345", "server address <HOSTNAME>:<PORT>")
	numRecords := flag.Int("n", 10000, "Num DB Records")
	recordSize := flag.Int("r", 1000, "Record size in bytes")
	numWorkers := flag.Int("w", 2, "Num workers")
	pirTypeStr := flag.String("t", "punc", fmt.Sprintf("PIR type: [%s]", strings.Join(b.PirTypeStrings(), "|")))
	hintProf := flag.String("hintprof", "", "Profile Server.Hint filename")
	answerProf := flag.String("answerprof", "", "Profile Server.Answer filename")

	flag.Parse()

	fmt.Printf("Connecting to %s...", *serverAddr)
	remote, err := rpc.DialHTTP("tcp", *serverAddr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	proxyLeft := b.NewPirRpcProxy(remote)
	proxyRight := b.NewPirRpcProxy(remote)
	client := b.NewPirClientUpdatable(b.RandSource(), [2]b.PirServer{proxyLeft, proxyRight})

	fmt.Printf("Setting up remote DB...")
	var none int

	pirType, err := b.PirTypeString(*pirTypeStr)
	if err != nil {
		log.Fatalf("Bad PirType: %s", *pirTypeStr)
	}
	err = proxyLeft.SetPIRType(pirType, &none)
	err = proxyLeft.ResetDBDimensions(b.DBDimensions{*numRecords, *recordSize}, &none)
	if err != nil {
		log.Fatalf("Failed to ResetDBDimensions: %s\n", err)
	}

	cachedVals := make(map[int]b.Row)
	cachedIndices := make([]int, 0)
	cachedKeys := make([]uint32, 0)
	for i := 0; i < NumDifferentReads; i++ {
		idx := i % *numRecords
		cachedIndices = append(cachedIndices, idx)
		cachedKeys = append(cachedKeys, rand.Uint32())
		cachedVals[idx] = make([]byte, *recordSize)
		rand.Read(cachedVals[idx])

		err = proxyLeft.SetRecordValue(
			b.RecordIndexVal{Index: idx, Key: cachedKeys[idx], Value: cachedVals[idx]},
			&none)
		if err != nil {
			log.Fatalf("Failed to SetRecordValue: %s\n", err)
		}
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
		idx := cachedIndices[rand.Intn(len(cachedIndices))]
		readVal, err := client.Read(int(cachedKeys[idx]))
		if err != nil {
			log.Fatalf("Failed to read index %d: %s", i, err)
		}
		if !reflect.DeepEqual(cachedVals[idx], readVal) {
			log.Fatalf("Mismatching record value at index %d", idx)
		}
	}
	proxyLeft.ShouldRecord = false
	proxyRight.ShouldRecord = false
	fmt.Printf("(%d #cached) [OK]\n", len(proxyRight.QueryReqs))

	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(1 * time.Second)

	if len(*answerProf) > 0 {
		err = proxyLeft.StartCpuProfile(0, &none)
		if err != nil {
			log.Fatalf("Failed to StartCpuProfile: %s\n", err)
		}
	}

	for i := 0; i < *numWorkers; i++ {
		go func() {
			for {
				idx := rand.Intn(len(proxyRight.QueryReqs))
				var queryResp b.QueryResp
				err := proxyRight.Answer(proxyRight.QueryReqs[idx], &queryResp)
				if err != nil {
					log.Fatalf("Failed to replay query number %d: %s\n", idx, err)
				}
				if !reflect.DeepEqual(proxyRight.QueryResps[idx], queryResp) {
					log.Fatalf("Mismatching response in query number %d", idx)
				}
				counter.Incr(1)
			}
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
		os.Exit(0)
	}()

	for {
		fmt.Printf("\rCurrent rate: %d QPS", counter.Rate())
	}
}
