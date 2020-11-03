package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type pirClientMatrix struct {
	height int
	width  int
	rowLen int

	randSource *rand.Rand
}

type pirMatrix struct {
	numRows int
	height  int
	width   int
	rowLen  int
	flatDb  []byte
}

func getHeightWidth(nRows int, rowLen int) (int, int) {
	// h^2 = n * rowlen
	width := int(math.Ceil(math.Sqrt(float64(nRows*rowLen)) / float64(rowLen)))
	height := (nRows-1)/width + 1

	return width, height
}

func NewPirServerMatrix(data []Row) PirServer {
	if len(data) < 1 {
		panic("Database must contain at least one row")
	}

	rowLen := len(data[0])
	flatDb := make([]byte, rowLen*len(data))

	for i, v := range data {
		if len(v) != rowLen {
			panic("Database rows must all be of the same length")
		}

		copy(flatDb[i*rowLen:], v[:])
	}

	width, height := getHeightWidth(len(data), rowLen)
	return &pirMatrix{
		numRows: len(data),
		rowLen:  rowLen,
		flatDb:  flatDb,
		height:  height,
		width:   width,
	}
}

func (s pirMatrix) matVecProduct(bitVector []bool) []byte {
	out := make([]byte, s.width*s.rowLen)

	cnt := 0
	tableWidth := s.rowLen * s.width
	for j := 0; j < s.height; j++ {
		if bitVector[j] {
			start := tableWidth * j
			length := tableWidth
			if start+length >= len(s.flatDb) {
				length = len(s.flatDb) - start
			}
			xorInto(out[0:length], s.flatDb[start:start+length])
			cnt = cnt + tableWidth
		}
	}
	return out
}

func (s pirMatrix) Hint(req HintReq, resp *HintResp) error {
	*resp = HintResp{
		NumRows: s.numRows,
		RowLen:  s.rowLen,
	}
	return nil
}

func (s *pirMatrix) Answer(q QueryReq, resp *QueryResp) error {
	*resp = QueryResp{Answer: s.matVecProduct(q.BitVector)}
	return nil
}

func (s *pirMatrix) NumRows(none int, out *int) error {
	*out = s.numRows
	return nil
}

func (s *pirMatrix) GetRow(idx int, row *RowIndexVal) error {
	if idx < 0 || idx >= s.numRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.numRows)
	}
	row.Value = s.flatDb[idx*s.rowLen : (idx+1)*s.rowLen]
	row.Index = idx
	row.Key = uint32(idx)
	return nil
}

func NewPirClientMatrix(source *rand.Rand) *pirClientMatrix {
	return &pirClientMatrix{randSource: source}
}

func (c *pirClientMatrix) initHint(resp *HintResp) error {
	c.rowLen = resp.RowLen
	c.width, c.height = getHeightWidth(resp.NumRows, c.rowLen)
	return nil
}

func (c *pirClientMatrix) query(idx int) ([]QueryReq, ReconstructFunc) {
	rowNum := idx / c.width
	colNum := idx % c.width
	queries := make([]QueryReq, 2)
	queries[Left].BitVector = make([]bool, c.height)
	queries[Right].BitVector = make([]bool, c.height)
	for i := 0; i < c.height; i++ {
		queries[Left].BitVector[i] = (c.randSource.Uint64()&1 == 0)
		queries[Right].BitVector[i] = queries[Left].BitVector[i] != (i == rowNum)
	}

	return queries, func(resps []QueryResp) (Row, error) {
		return c.reconstruct(colNum, resps)
	}
}

func (c *pirClientMatrix) dummyQuery() []QueryReq {
	q, _ := c.query(0)
	return q
}

func (c *pirClientMatrix) reconstruct(colNum int, resp []QueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out[c.rowLen*colNum : (c.rowLen * (colNum + 1))], nil
}
