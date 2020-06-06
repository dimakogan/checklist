package boosted

import (
	"fmt"
	"math/rand"
)

type PIRServerStub struct {
	db []Row

	// Simulated work params.
	numReadsOnHint   int
	numReadsOnAnswer int

	randSource *rand.Rand
}

func (s PIRServerStub) fakeProbes(n int) {
	totalSum := 0
	for i := 0; i < n; i++ {
		st := s.db[s.randSource.Intn(len(s.db))]
		for j := 0; j < len(st); j++ {
			totalSum += int(st[j])
		}
	}
}

func (s PIRServerStub) Hint(*HintReq, *HintResp) error {
	s.fakeProbes(s.numReadsOnHint)
	return nil
}

func (s PIRServerStub) Answer(q *QueryReq, resp *QueryResp) error {
	s.fakeProbes(s.numReadsOnAnswer)
	resp.Val = s.db[q.Index]
	return nil
}

type pirClientStub struct {
}

func NewPIRServerStub(db []Row, numReadsOnHint int, numReadsOnAnswer int, randSource *rand.Rand) PIRServer {
	return PIRServerStub{
		db: db,
		numReadsOnHint: numReadsOnHint,
		numReadsOnAnswer: numReadsOnAnswer,
		randSource: randSource,
	}
}

func NewPirClientStub() PIRClient {
	return pirClientStub{}
}

func (c pirClientStub) RequestHintN(nHints int) (*HintReq, error) {
  return c.RequestHint()
}

func (c pirClientStub) RequestHint() (*HintReq, error) {
	return &HintReq{}, nil
}

func (c pirClientStub) InitHint(*HintResp) error {
	return nil
}

func (c pirClientStub) Query(i int) ([]*QueryReq, error) {
	return []*QueryReq{&QueryReq{Index: i}}, nil
}

func (c pirClientStub) Reconstruct(resp []*QueryResp) (Row, error) {
	if len(resp) != 1 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(resp))
	}
	return resp[0].Val, nil
}
