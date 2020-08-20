package boosted

import (
	"fmt"
	"math"
)

// One database row.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct {
	NumHints  int
	SetSize   int
	SetGenKey []byte

	// For PirPerm trial
	PartitionKey []byte

	// For PirMatrix
	BitVector []bool
}

//HintResp is a response to a hint request.
type HintResp struct {
	Hints []Row
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	PuncturedSet SuccinctSet
	ExtraElem    int

	// For PirMatrix
	BitVector []bool

	// For PirErasure
	BatchReqs []QueryReq

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
	Hint(req *HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *QueryResp) error
}

type PirClient interface {
	Init() error
	Read(i int) (Row, error)
}

type ReconstructFunc func(resp []QueryResp) (Row, error)

type pirClientImpl interface {
	requestHint() (*HintReq, error)
	initHint(resp *HintResp) error
	query(i int) ([]QueryReq, ReconstructFunc)
}

type pirClient struct {
	impl    pirClientImpl
	servers [2]PirServer
}

func NewPIRClient(impl pirClientImpl, servers [2]PirServer) PirClient {
	return pirClient{impl: impl, servers: servers}
}

func (c pirClient) Init() error {
	hintReq, err := c.impl.requestHint()
	if err != nil {
		return err
	}
	var hintResp HintResp
	err = c.servers[Left].Hint(hintReq, &hintResp)
	if err != nil {
		return err
	}
	return c.impl.initHint(&hintResp)
}

func (c pirClient) Read(i int) (Row, error) {
	queryReq, reconstructFunc := c.impl.query(i)
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
		panic("Database must contain at least one row")
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

func numRecordsToUnivSizeBits(nRecords int) int {
	// Round univsize to next power of 4
	return ((int(math.Log2(float64(nRecords)))-1)/2 + 1) * 2
}
