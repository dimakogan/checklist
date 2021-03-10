package boosted

import (
	"math"
	"math/rand"

	"github.com/dimakogan/boosted-pir/psetggm"
	"github.com/dimakogan/dpf-go/dpf"
)

type pirDPFClient struct {
	nRows int

	randSource *rand.Rand
}

type pirDPFServer struct {
	*staticDB
}

func NewPIRDPFServer(db *staticDB) *pirDPFServer {
	if db.numRows < 1 {
		panic("Database must contain at least one row")
	}

	return &pirDPFServer{db}
}

func (s pirDPFServer) matVecProduct(bitVector []byte) []byte {
	out := make(Row, s.rowLen)
	if s.rowLen == 32 {
		psetggm.XorHashesByBitVector(s.flatDb, bitVector, out)
	} else {
		var j uint
		for j = 0; j < uint(s.numRows); j++ {
			if ((1 << (j % 8)) & bitVector[j/8]) != 0 {
				xorInto(out, s.flatDb[j*uint(s.rowLen):(j+1)*uint(s.rowLen)])
			}
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

func NewPIRDPFClient(source *rand.Rand) *pirDPFClient {
	return &pirDPFClient{randSource: source}
}

func (c *pirDPFClient) InitHint(resp *HintResp) error {
	c.nRows = resp.NumRows
	return nil
}

func (c *pirDPFClient) Query(idx int) ([]QueryReq, ReconstructFunc) {
	queries := make([]QueryReq, 2)
	numBits := uint64(math.Ceil(math.Log2(float64(c.nRows))))
	queries[Left].DPFkey, queries[Right].DPFkey = dpf.Gen(uint64(idx), numBits)

	return queries, c.reconstruct
}

func (c *pirDPFClient) dummyQuery() []QueryReq {
	q, _ := c.Query(0)
	return q
}

func (c *pirDPFClient) reconstruct(resp []QueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out, nil
}
