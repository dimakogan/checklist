package main

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
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
	pirType := flag.String("t", "punc", "PIR Type: [punc|matrix]")

	flag.Parse()

	fmt.Printf("Connecting to %s...", *serverAddr)
	remote, err := rpc.DialHTTP("tcp", *serverAddr)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	fmt.Printf("[OK]\n")

	proxyLeft := b.NewPirRpcProxy(remote)
	proxyRight := b.NewPirRpcProxy(remote)
	var client b.PirClient
	switch *pirType {
	case "punc":
		client = b.NewPIRClient(b.NewPirClientPunc(b.RandSource(), *numRecords), [2]b.PirServer{proxyLeft, proxyRight})
	case "matrix":
		client = b.NewPIRClient(b.NewPirClientMatrix(b.RandSource(), *numRecords, *recordSize), [2]b.PirServer{proxyLeft, proxyRight})
	}

	fmt.Printf("Setting up remote DB...")
	var none int
	err = remote.Call("PirRpcServer.SetPIRType", *pirType, &none)
	err = remote.Call("PirRpcServer.SetDBDimensions", b.DBDimensions{*numRecords, *recordSize}, &none)
	if err != nil {
		log.Fatalf("Failed to SetDBDimensions: %s\n", err)
	}

	cachedVals := make(map[int]b.Row)
	cachedIndices := make([]int, 0)
	for i := 0; i < NumDifferentReads; i++ {
		idx := i % *numRecords
		cachedIndices = append(cachedIndices, idx)
		cachedVals[idx] = make([]byte, *recordSize)
		rand.Read(cachedVals[idx])

		err = remote.Call(
			"PirRpcServer.SetRecordValue",
			b.RecordIndexVal{Index: idx, Value: cachedVals[idx]},
			&none)
		if err != nil {
			log.Fatalf("Failed to SetRecordValue: %s\n", err)
		}
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Obtaining hint (this may take a while)...")
	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Caching responses...")
	proxyLeft.ShouldRecord = true
	proxyRight.ShouldRecord = true
	for i := 0; i < NumDifferentReads; i++ {
		idx := cachedIndices[rand.Intn(len(cachedIndices))]
		readVal, err := client.Read(idx)
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

	for {
		fmt.Printf("\rCurrent rate: %d QPS", counter.Rate())
	}
}
