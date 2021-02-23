package main

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"

	b "github.com/dimakogan/boosted-pir"
)

type answerStresser struct {
	reqs  []b.QueryReq
	resps []b.QueryResp
}

func initAnswerStresser(config *b.Configurator, proxy *b.PirRpcProxy) *answerStresser {
	for i := 0; i < NumDifferentReads; i++ {
		idx := i % config.NumRows
		value := make([]byte, config.RowLen)
		rand.Read(value)
		config.PresetRows = append(config.PresetRows, b.RowIndexVal{
			Index: idx,
			Key:   rand.Uint32(),
			Value: value})
	}

	err := proxy.Configure(config.TestConfig, nil)
	if err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}

	client := b.NewPirClientUpdatable(b.RandSource(), config.PirType, [2]b.PirUpdatableServer{proxy, proxy})
	client.CallAsync = false

	fmt.Printf("Obtaining hint (this may take a while)...")
	err = client.Init()
	if err != nil {
		log.Fatalf("Failed to Initialize client: %s\n", err)
	}
	fmt.Printf("[OK]\n")

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
	reqs := make([]b.QueryReq, 0, len(ireqs))
	resps := make([]b.QueryResp, 0, len(ireqs))
	for i := range ireqs {
		reqs = append(reqs, ireqs[i].ReqBody.(b.QueryReq))
		resps = append(resps, *(ireqs[i].RespBody.(*b.QueryResp)))
	}
	if len(reqs) == 0 {
		log.Fatalf("Failed to cache any requests")
	}
	fmt.Printf("(%d #cached) [OK]\n", len(reqs))

	return &answerStresser{reqs: reqs, resps: resps}
}

func (s *answerStresser) request(proxy *b.PirRpcProxy) error {
	idx := rand.Intn(len(s.reqs))
	var queryResp b.QueryResp
	err := proxy.Answer(s.reqs[idx], &queryResp)
	if err != nil {
		return fmt.Errorf("Failed to replay query number %d to server: %s", idx, err)
	}
	if !reflect.DeepEqual(s.resps[idx], queryResp) {
		return fmt.Errorf("Mismatching response in query number %d", idx)
	}
	return nil
}
