package boosted

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
)

var SecParam = flag.Int("secParam", 128, "Security Parameter (in bits)")

//go:generate enumer -type=PirType
type PirType int

const (
	Matrix PirType = iota
	Punc
	Perm
)

// One database row.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct {
	RandSeed int64

	// For PirUpdatable
	LatestKeyTimestamp int
}

type TimedRow struct {
	Timestamp int
	Key       uint32
	Delete    bool
	data      Row
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

	// For updatable PIR
	EndTimestamp        int
	TimedKeys           []TimedRow
	ShouldDeleteHistory bool

	BatchResps []HintResp
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	PuncturedSet SuccinctSet
	ExtraElem    int

	// For PirMatrix
	BitVector []bool

	// For PirErasure
	BatchReqs []QueryReq

	// For PirUpdatable
	LatestKeyTimestamp int

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

type PirDB interface {
	AddRows(keys []uint32, vals []Row)
	DeleteRows(keys []uint32)
}

type PirServer interface {
	Hint(req HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *QueryResp) error
}

type PirClient interface {
	Init() error
	Read(i int) (Row, error)
}

type ReconstructFunc func(resp []QueryResp) (Row, error)

type pirClientImpl interface {
	initHint(resp *HintResp) error
	query(i int) ([]QueryReq, ReconstructFunc)
	dummyQuery() []QueryReq
}

type pirClient struct {
	impl       pirClientImpl
	servers    [2]PirServer
	randSource *rand.Rand
}

func NewPirServerByType(pirType PirType, randSrc *rand.Rand, db []Row) PirServer {
	switch pirType {
	case Matrix:
		return NewPirServerMatrix(randSrc, db)
	case Punc:
		return NewPirServerPunc(randSrc, db)
	case Perm:
		return NewPirPermServer(db)
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}

func NewPirClientByType(pirType PirType, randSrc *rand.Rand) pirClientImpl {
	switch pirType {
	case Matrix:
		return NewPirClientMatrix(randSrc)
	case Punc:
		return NewPirClientPunc(randSrc)
	case Perm:
		return NewPirPermClient(randSrc)
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}

func NewPIRClient(impl pirClientImpl, source *rand.Rand, servers [2]PirServer) PirClient {
	return pirClient{impl: impl, servers: servers, randSource: source}
}

func (c pirClient) Init() error {
	hintReq := HintReq{RandSeed: int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}
	return c.impl.initHint(&hintResp)
}

func (c pirClient) Read(i int) (Row, error) {
	queryReq, reconstructFunc := c.impl.query(i)
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %d", i)
	}
	responses := make([]QueryResp, 2)
	err := c.servers[Left].Answer(queryReq[Left], &responses[Left])
	if err != nil {
		return nil, err
	}

	err = c.servers[Right].Answer(queryReq[Right], &responses[Right])
	if err != nil {
		return nil, err
	}
	return reconstructFunc(responses)
}

func flattenDb(data []Row) []byte {
	if len(data) < 1 {
		return []byte{}
	}

	rowLen := len(data[0])
	flatDb := make([]byte, rowLen*len(data))

	for i, v := range data {
		if len(v) != rowLen {
			fmt.Printf("Got row[%v] %v %v\n", i, len(v), rowLen)
			panic("Database rows must all be of the same length")
		}

		copy(flatDb[i*rowLen:], v[:])
	}
	return flatDb
}

func xorRowsFlatSlice(flatDb []byte, rowLen int, rows Set, out []byte) {
	nRows := len(flatDb) / rowLen
	for _, row := range rows {
		if row >= nRows {
			continue
		}
		xorInto(out, flatDb[rowLen*row:rowLen*(row+1)])
	}
}

func numRowsToUnivSizeBits(nRows int) int {
	// Round univsize to next power of 4
	return ((int(math.Log2(float64(nRows)))-1)/2 + 1) * 2
}

func PirTypeStrings() []string {
	vals := PirTypeValues()
	strs := make([]string, len(vals))
	for i, val := range vals {
		strs[i] = val.String()
	}
	return strs
}
