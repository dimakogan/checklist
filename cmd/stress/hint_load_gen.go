package main

import (
	"fmt"
	"math"
	"math/rand"

	. "github.com/dimakogan/boosted-pir"
)

type hintLoadGen struct {
	numRows int
	sizes   []int
	probs   []float64
	pirType PirType
}

func initHintLoadGen(config *Config, proxy *PirRpcProxy) *hintLoadGen {
	sizes := NewPirClientUpdatable(RandSource(), config.PirType, [2]PirUpdatableServer{proxy, proxy}).LayersMaxSize(config.NumRows)
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

func (s *hintLoadGen) request(proxy *PirRpcProxy) error {
	layerSize := s.randSize()
	firstRow := rand.Intn(s.numRows - layerSize + 1)
	hintReq := HintReq{
		RandSeed:        42,
		DefragTimestamp: math.MaxInt32,
		FirstRow:        firstRow,
		NumRows:         layerSize,
		PirType:         s.pirType,
	}
	//fmt.Printf("Using size: %d\n", layerSize)
	var hintResp HintResp
	err := proxy.Hint(hintReq, &hintResp)
	if err != nil {
		return fmt.Errorf("Failed to replay hint request %v, %s", hintReq, err)
	}
	if hintResp.NumRows != layerSize {
		fmt.Printf("%+v\n", hintResp)
		return fmt.Errorf("Failed to replay hint request %v , mismatching hint num rows, expected: %d, got: %d", hintReq, layerSize, hintResp.NumRows)
	}
	return nil
}
