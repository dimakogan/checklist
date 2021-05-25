package pir

import (
	"math/rand"
)

const (
	// SecParam is the security parameter in bits.
	SecParam = 128
)

const Left int = 0
const Right int = 1

// One database row.
type Row []byte

// HintReq is a request for a hint from a client to a server.

type HintReq interface {
	Process(db StaticDB) (HintResp, error)
}

type HintResp interface {
	InitClient(randSrc *rand.Rand) Client
	NumRows() int
}

type Client interface {
	Query(i int) ([]QueryReq, ReconstructFunc)
	DummyQuery() []QueryReq
	StateSize() (bitsPerKey, fixedBytes int)
}

//QueryReq is a PIR query from a client to a server.
type QueryReq interface {
	Process(db StaticDB) (interface{}, error)
}

type ReconstructFunc func(resp []interface{}) (Row, error)
