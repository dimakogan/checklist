package main

import (
	"fmt"
)

type pirServerStub struct {
	db []string
}

func newPirServerStub(db []string) PIRServer {
	return pirServerStub{db: db}
}

func (s pirServerStub) Hint(*HintReq) (*HintResp, error) {
	return nil, nil
}

func (s pirServerStub) Answer(q *QueryReq) (*QueryResp, error) {
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
