package boosted

import (
	"fmt"
	"math"
)

// One database row.
type Row []byte

//HintReq is a request for a hint from a client to a server.
type HintReq struct {
	Sets          []SetKey
	AuxRecordsSet SetKey

	// For PirPerm trial
	NumHints     int
	PartitionKey []byte
}

//HintResp is a response to a hint request.
type HintResp struct {
	Hints      []Row
	AuxRecords map[int]Row
}

//QueryReq is a PIR query from a client to a server.
type QueryReq struct {
	PuncturedSet SetKey

	// Debug & testing.
	Index int
}

//QueryResp is a response to a PIR query.
type QueryResp struct {
	Answer Row

	// For PirPerm trial
	Values []Row

	// Debug & testing
	Val Row
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
