package main

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"

	. "github.com/dimakogan/boosted-pir"
)

type answerLoadGen struct {
	numRows int
	reqs    []QueryReq
	resps   []QueryResp
}

func initAnswerLoadGen(config *Config) *answerLoadGen {
	fmt.Printf("Connecting to %s (TLS: %t)...", config.ServerAddr, config.UseTLS)
	proxy, err := NewPirRpcProxy(config.ServerAddr, config.UseTLS, config.UsePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	defer proxy.Close()
	fmt.Printf("[OK]\n")

	totalNumRows := config.NumRows
	config.NumRows = config.NumRows * 2 / 3

	for i := 0; i < NumDifferentReads; i++ {
		idx := i % config.NumRows
		value := make([]byte, config.RowLen)
		rand.Read(value)
		config.PresetRows = append(config.PresetRows, RowIndexVal{
			Index: idx,
			Key:   rand.Uint32(),
			Value: value})
	}

	fmt.Printf("Setting up remote DB...")
	if err := proxy.Configure(config.TestConfig, nil); err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}
	fmt.Printf("[OK]\n")

	client := NewPirClientUpdatable(RandSource(), config.PirType, [2]PirUpdatableServer{proxy, proxy})
	client.CallAsync = false

	fmt.Printf("Obtaining hint (this may take a while)...")
	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}
	fmt.Printf("[OK]\n")

	fmt.Printf("Adding more rows...")
	for config.NumRows < totalNumRows {
		toAdd := ((totalNumRows-config.NumRows)/2 + 1)
		if err = proxy.AddRows(toAdd, nil); err != nil {
			log.Fatalf("failed to add %d rows: %s", toAdd, err)
		}
		config.NumRows += toAdd
		if err = client.Update(); err != nil {
			log.Fatalf("failed to update hint after adding %d rows: %s", toAdd, err)
		}
	}
	var numRows int
	if err = proxy.NumRows(0, &numRows); err != nil || numRows < config.NumRows*99/100 {
		log.Fatalf("Invalid number of rows on server: %d", numRows)
	}

	fmt.Printf("[OK] (num rows: %d)\n", config.NumRows)

	fmt.Printf("Caching responses...")
	proxy.StartRecording()
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
	ireqs := proxy.StopRecording()
	reqs := make([]QueryReq, 0, len(ireqs))
	resps := make([]QueryResp, 0, len(ireqs))
	for i := range ireqs {
		reqs = append(reqs, ireqs[i].ReqBody.(QueryReq))
		resps = append(resps, *(ireqs[i].RespBody.(*QueryResp)))
	}
	if len(reqs) == 0 {
		log.Fatalf("Failed to cache any requests")
	}
	fmt.Printf("(%d #cached) [OK]\n", len(reqs))

	return &answerLoadGen{numRows: numRows, reqs: reqs, resps: resps}
}

func (s *answerLoadGen) request(proxy *PirRpcProxy) error {
	idx := rand.Intn(len(s.reqs))
	var queryResp QueryResp
	err := proxy.Answer(s.reqs[idx], &queryResp)
	if err != nil {
		return fmt.Errorf("Failed to replay query number %d to server: %s", idx, err)
	}
	if !reflect.DeepEqual(s.resps[idx], queryResp) {
		return fmt.Errorf("Mismatching response in query number %d", idx)
	}
	return nil
}

func (gen *answerLoadGen) debugStr() string {
	return ""
}

func (gen *answerLoadGen) reqRate() int {
	return 1
}
