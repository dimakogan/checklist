package main

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"time"

	b "github.com/dimakogan/boosted-pir"

	"github.com/paulbellamy/ratecounter"

	"log"
	"net/rpc"
)

// Number of different records to read to avoid caching effects.
var NumDifferentReads = 100

func main() {
	args := os.Args
	if len(args) < 4 {
		panic(fmt.Sprintf("Usage: %s <SERVER-ADDR> <NUM-DB-RECORDS> <RECORD-SIZE> <NUM-WORKERS>", args[0]))
	}
	numRecords, err := strconv.Atoi(args[2])
	if err != nil {
		panic(fmt.Sprintf("Invalid NUM-DB-RECORDS: %s", args[2]))
	}

	recordSize, err := strconv.Atoi(args[3])
	if err != nil {
		panic(fmt.Sprintf("Invalid RECORD-SIZE: %s", args[3]))
	}

	numWorkers, err := strconv.Atoi(args[4])
	if err != nil {
		panic(fmt.Sprintf("Invalid NUM-WORKERS: %s", args[4]))
	}

	remote, err := rpc.DialHTTP("tcp", args[1])
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	proxyLeft := b.NewPirRpcProxy(remote)
	proxyRight := b.NewPirRpcProxy(remote)
	client := b.NewPirClientPunc(b.RandSource(), numRecords, [2]b.PirServer{proxyLeft, proxyRight})
	proxyLeft.ShouldRecord = true
	proxyRight.ShouldRecord = true

	var none int
	err = remote.Call("PirRpcServer.SetDBDimensions", b.DBDimensions{numRecords, recordSize}, &none)
	if err != nil {
		log.Fatalf("Failed to SetDBDimensions: %s\n", err)
	}

	cachedVals := make(map[int]b.Row)
	cachedIndices := make([]int, 0)
	for i := 0; i < NumDifferentReads; i++ {
		idx := i % numRecords
		cachedIndices = append(cachedIndices, idx)
		cachedVals[idx] = make([]byte, recordSize)
		rand.Read(cachedVals[idx])

		err = remote.Call(
			"PirRpcServer.SetRecordValue",
			b.RecordIndexVal{Index: idx, Value: cachedVals[idx]},
			&none)
		if err != nil {
			log.Fatalf("Failed to SetRecordValue: %s\n", err)
		}
	}

	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}

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
	proxyRight.ShouldRecord = false
	for i := range proxyRight.QueryReqs {
		queryResps := make([]b.QueryResp, len(proxyRight.QueryReqs[i]))
		err = proxyRight.AnswerBatch(proxyRight.QueryReqs[i], &queryResps)
		if err != nil {
			log.Fatalf("Failed to replay query number %d: %s\n", i, err)
		}
		if !reflect.DeepEqual(proxyRight.QueryResps[i], queryResps) {
			log.Fatalf("Mismatching response in query number %d", i)
		}
	}

	fmt.Printf("Initial response verification OK (#cached vals: %d, requests: %d, responses: %d)\n",
		len(cachedIndices), len(proxyRight.QueryReqs), len(proxyRight.QueryResps))

	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(1 * time.Second)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				idx := rand.Intn(len(proxyRight.QueryReqs))
				queryResps := make([]b.QueryResp, len(proxyRight.QueryReqs[idx]))
				err := proxyRight.AnswerBatch(proxyRight.QueryReqs[idx], &queryResps)
				if err != nil {
					log.Fatalf("Failed to replay query number %d: %s\n", idx, err)
				}
				if !reflect.DeepEqual(proxyRight.QueryResps[idx], queryResps) {
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
