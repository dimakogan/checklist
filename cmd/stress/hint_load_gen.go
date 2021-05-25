package main

import (
	"fmt"
	"log"
	"math/rand"

	. "checklist/driver"
	"checklist/pir"
	"checklist/updatable"
)

type hintLoadGen struct {
	numRows int
	sizes   []int
	probs   []float64
	pirType pir.PirType
}

func initHintLoadGen(config *Config) *hintLoadGen {
	fmt.Printf("Connecting to %s (TLS: %t)...", config.ServerAddr, config.UseTLS)
	proxy, err := NewRpcProxy(config.ServerAddr, config.UseTLS, config.UsePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	defer proxy.Close()
	fmt.Printf("[OK]\n")

	fmt.Printf("Setting up remote DB...")
	if err := proxy.Configure(config.TestConfig, nil); err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}
	var numRows int
	if err = proxy.NumRows(0, &numRows); err != nil || numRows < config.NumRows*99/100 {
		log.Fatalf("Invalid number of rows on server: %d", numRows)
	}

	fmt.Printf("[OK] (num rows: %d)\n", config.NumRows)

	sizes := updatable.NewWaterfallClient(pir.RandSource(), config.PirType).LayersMaxSize(config.NumRows)
	probs := make([]float64, len(sizes))
	probs[len(probs)-1] = 1.0
	overflowSize := 2 * sizes[len(sizes)-1]
	for l := len(sizes) - 2; l >= 0; l-- {
		// Number of time this layer gots activated before it overflows
		c := sizes[l] / overflowSize
		probs[l] = probs[l+1] * (float64(c) / float64(c+1))
		overflowSize = (c + 1) * overflowSize
	}
	fmt.Printf("Using layer sizes %v with probabilities %v\n", sizes, probs)
	return &hintLoadGen{config.NumRows, sizes, probs, config.PirType}
}

func (s *hintLoadGen) randSize() int {
	p := rand.Float64()
	bucket := 0
	for p > s.probs[bucket] {
		bucket++
	}
	return s.sizes[bucket]
}

func (s *hintLoadGen) request(proxy *RpcProxy) error {
	layerSize := s.randSize()
	firstRow := rand.Intn(s.numRows - layerSize + 1)
	hintReq := updatable.UpdatableHintReq{
		FirstRow: firstRow,
		NumRows:  layerSize,
		Req:      pir.NewHintReq(rand.New(rand.NewSource(42)), s.pirType),
	}
	var hintResp pir.HintResp
	err := proxy.Hint(&hintReq, &hintResp)
	if err != nil {
		return fmt.Errorf("Failed to replay hint request %v, %s", hintReq, err)
	}
	if hintResp.NumRows() != layerSize {
		fmt.Printf("%+v\n", hintResp)
		return fmt.Errorf("Failed to replay hint request %v , mismatching hint num rows, expected: %d, got: %d", hintReq, layerSize, hintResp.NumRows())
	}
	return nil
}

func (gen *hintLoadGen) debugStr() string {
	return ""
}

func (gen *hintLoadGen) reqRate() int {
	return 1
}
