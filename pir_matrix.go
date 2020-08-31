package boosted

import (
	"math"
	"math/rand"
)

type pirClientMatrix struct {
	height  int
	width   int
	rowLen  int
	xorLeft []Row

	randSource *rand.Rand
}

type pirMatrix struct {
	height int
	width  int
	rowLen int
	flatDb []byte

	randSource *rand.Rand
}

func getHeightWidth(nRows int, rowLen int) (int, int) {
	// h^2 = n * rowlen
	height := int(math.Sqrt(float64(nRows * rowLen)))
	width := nRows / height
	return width, height
}

func NewPirServerMatrix(source *rand.Rand, data []Row) *pirMatrix {
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
		rowLen:     rowLen,
		flatDb:     flatDb,
		randSource: source,
		height:     height,
		width:      width,
	}
}

func (s pirMatrix) matVecProduct(bitVector []bool) []byte {
	out := make([]byte, s.width*s.rowLen)

	cnt := 0
	tableWidth := s.rowLen * s.width
	for j := 0; j < s.height; j++ {
		if bitVector[j] {
			xorInto(out, s.flatDb[tableWidth*j:(tableWidth*(j+1))])
		}
		cnt = cnt + tableWidth
	}
	return out
}

func (s pirMatrix) Hint(req HintReq, resp *HintResp) error {
	return nil
}

func (s *pirMatrix) Answer(q QueryReq, resp *QueryResp) error {
	*resp = QueryResp{Answer: s.matVecProduct(q.BitVector)}
	return nil
}

func NewPirClientMatrix(source *rand.Rand, nRows int, rowLen int) *pirClientMatrix {
	width, height := getHeightWidth(nRows, rowLen)
	return &pirClientMatrix{
		rowLen:     rowLen,
		randSource: source,
		height:     height,
		width:      width,
	}
}

func (c *pirClientMatrix) requestHint() (*HintReq, error) {
	return &HintReq{}, nil
}

func (c *pirClientMatrix) initHint(resp *HintResp) error {
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

func (c *pirClientMatrix) reconstruct(colNum int, resp []QueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out[c.rowLen*colNum : (c.rowLen * (colNum + 1))], nil
}
