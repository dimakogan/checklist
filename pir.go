package boosted

import (
	"fmt"
	"math/rand"

	"github.com/dimakogan/dpf-go/dpf"
)

const (
	// SecParam is the security parameter in bits.
	SecParam = 128
)

//go:generate enumer -type=PirType
type PirType int

const (
	None PirType = iota
	Matrix
	Punc
	Perm
	DPF
	NonPrivate
)

// One database row.
type Row []byte

// HintReq is a request for a hint from a client to a server.
type HintReq struct {
	// Random seed to be used by the server to generate the hint.
	RandSeed int64

	// Type of PIR to use.
	PirType PirType

	// For PirUpdatable
	DefragTimestamp int32
	FirstRow        int
	NumRows         int
}

//HintResp is a response to a hint request.
type HintResp struct {
	PirType   PirType
	NumRows   int
	RowLen    int
	SetSize   int
	SetGenKey []byte
	Hints     []Row
	IsMatrix  bool
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	PuncturedSet PuncturedSet
	ExtraElem    int

	// For PirMatrix
	BitVector []bool

	// For PirErasure
	BatchReqs []QueryReq

	// For PirDPF
	DPFkey dpf.DPFkey

	// For PirUpdatable
	FirstRow, NumRows int32
	PirType           PirType

	// Debug & testing.
	Index int
}

//QueryResp is a response to a PIR query.
type QueryResp struct {
	Answer    Row
	ExtraElem Row

	// For PirPerm trial
	Values []Row

	// For PirErasure
	BatchResps []QueryResp

	// Debug & testing
	Val Row
}

type PirServer interface {
	Hint(req HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *QueryResp) error
}

type ReconstructFunc func(resp []QueryResp) (Row, error)

type PIRClient interface {
	InitHint(resp *HintResp) error
	Query(i int) ([]QueryReq, ReconstructFunc)
	dummyQuery() []QueryReq
}

func NewPirServerByType(pirType PirType, db *staticDB) PirServer {
	switch pirType {
	case Matrix:
		return NewPirServerMatrix(db)
	case Punc:
		return NewPirServerPunc(db)
	case DPF:
		return NewPIRDPFServer(db)
	case NonPrivate:
		return NewPirServerNonPrivate(db)
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}

func NewPirClientByType(pirType PirType, randSrc *rand.Rand) PIRClient {
	switch pirType {
	case Matrix:
		return NewPirClientMatrix(randSrc)
	case Punc:
		return NewPirClientPunc(randSrc)
	case DPF:
		return NewPIRDPFClient(randSrc)
	case NonPrivate:
		return NewPirClientNonPrivate()
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}
