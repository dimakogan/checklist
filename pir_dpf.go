package boosted

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/dimakogan/dpf-go/dpf"
)

type pirDPFClient struct {
	numRows, rowLen int

	randSource *rand.Rand
}

type pirDPFServer struct {
	numRows, rowLen int
	flatDb          []byte
}

func NewPIRDPFServer(data []Row) PirServer {
	if len(data) < 1 {
		panic("Database must contain at least one row")
	}

	rowLen := len(data[0])
	flatDb := flattenDb(data)

	return &pirDPFServer{
		numRows: len(data),
		rowLen:  rowLen,
		flatDb:  flatDb,
	}
}

func (s pirDPFServer) matVecProduct(bitVector []byte) []byte {
	out := make(Row, s.rowLen)

	var j uint
	for j = 0; j < uint(s.numRows); j++ {
		if ((1 << (j % 8)) & bitVector[j/8]) != 0 {
			xorInto(out, s.flatDb[j*uint(s.rowLen):(j+1)*uint(s.rowLen)])
		}
	}
	return out
}

func (s pirDPFServer) Hint(req HintReq, resp *HintResp) error {
	*resp = HintResp{
		NumRows: s.numRows,
		RowLen:  s.rowLen,
	}
	return nil
}

func (s *pirDPFServer) Answer(q QueryReq, resp *QueryResp) error {
	bitVec := dpf.EvalFull(q.DPFkey, uint64(math.Ceil(math.Log2(float64(s.numRows)))))
	*resp = QueryResp{Answer: s.matVecProduct(bitVec)}
	return nil
}

func (s *pirDPFServer) GetRow(idx int, row *RowIndexVal) error {
	if idx < 0 || idx >= s.numRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.numRows)
	}
	row.Value = s.flatDb[idx*s.rowLen : (idx+1)*s.rowLen]
	row.Index = idx
	row.Key = uint32(idx)
	return nil
}

func NewPIRDPFClient(source *rand.Rand) *pirDPFClient {
	return &pirDPFClient{randSource: source}
}

func (c *pirDPFClient) initHint(resp *HintResp) error {
	c.rowLen = resp.RowLen
	c.numRows = resp.NumRows
	return nil
}

func (c *pirDPFClient) query(idx int) ([]QueryReq, ReconstructFunc) {
	queries := make([]QueryReq, 2)
	numBits := uint64(math.Ceil(math.Log2(float64(c.numRows))))
	queries[Left].DPFkey, queries[Right].DPFkey = dpf.Gen(uint64(idx), numBits)

	return queries, c.reconstruct
}

func (c *pirDPFClient) dummyQuery() []QueryReq {
	q, _ := c.query(0)
	return q
}

func (c *pirDPFClient) reconstruct(resp []QueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out, nil
}
