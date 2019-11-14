package main

import (
	"fmt"
	"math/rand"
)

type pirServerStub struct {
	db []string

	// Simulated work params.
	numReadsOnHint   int
	numReadsOnAnswer int

	randSource *rand.Rand
}

func (s pirServerStub) fakeProbes(n int) {
	totalSum := 0
	for i := 0; i < n; i++ {
		st := s.db[s.randSource.Intn(len(s.db))]
		for j := 0; j < len(st); j++ {
			totalSum += int(st[j])
		}
	}
}

func (s pirServerStub) Hint(*HintReq) (*HintResp, error) {
	s.fakeProbes(s.numReadsOnHint)
	return nil, nil
}

func (s pirServerStub) Answer(q *QueryReq) (*QueryResp, error) {
	s.fakeProbes(s.numReadsOnAnswer)
	return &QueryResp{Val: s.db[q.Index]}, nil
}

type pirClientStub struct {
}

func newPirClientStub() PIRClient {
	return pirClientStub{}
}

func (c pirClientStub) RequestHint() (*HintReq, error) {
	return nil, nil
}

func (c pirClientStub) InitHint(*HintResp) error {
	return nil
}

func (c pirClientStub) Query(i int) ([]*QueryReq, error) {
	return []*QueryReq{&QueryReq{Index: i}}, nil
}

func (c pirClientStub) Reconstruct(resp []*QueryResp) (string, error) {
	if len(resp) != 1 {
		return "", fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(resp))
	}
	return resp[0].Val, nil
}
