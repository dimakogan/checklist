package boosted

import (
	"fmt"
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

type pirMatrixHintReq []bool
type pirMatrixHintResp []byte

func getHeightWidth(nRows int, rowLen int) (int, int) {
	// h^2 = n * rowlen
	height := int(math.Sqrt(float64(nRows * rowLen)))
	width := nRows / height
	return width, height
}

func NewPirServerMatrix(source *rand.Rand, data []Row, hintStrategy int) *pirMatrix {
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

func (s *pirMatrix) Hint(req pirMatrixHintReq, resp *pirMatrixHintResp) error {
	hint := make([]byte, s.width*s.rowLen)

	cnt := 0
	tableWidth := s.rowLen * s.width
	for j := 0; j < s.height; j++ {
		if req[j] {
			xorInto(hint, s.flatDb[tableWidth*j:(tableWidth*(j+1))])
		}
		cnt = cnt + tableWidth
	}

	*resp = hint
	return nil
}

func (s *pirMatrix) Answer(q *QueryReq, resp *QueryResp) error {
	return nil
}

func newPirClientMatrix(source *rand.Rand, nRows int, rowLen int) *pirClientMatrix {
	width, height := getHeightWidth(nRows, rowLen)
	return &pirClientMatrix{
		rowLen:     rowLen,
		randSource: source,
		height:     height,
		width:      width,
	}
}

func (c *pirClientMatrix) RequestHint() (*pirMatrixHintReq, error) {
	var hr pirMatrixHintReq
	hr = make([]bool, c.height)
	for i := 0; i < len(hr); i++ {
		hr[i] = (c.randSource.Uint64()&1 == 0)
	}

	return &hr, nil
}

func (c *pirClientMatrix) InitHint(resp *HintResp) error {
	return nil
}

func (c *pirClientMatrix) Query(i int) ([]*QueryReq, error) {
	return []*QueryReq{}, nil
}

func (c *pirClientMatrix) Reconstruct(resp []*QueryResp) (Row, error) {
	if len(resp) != 1 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(resp))
	}

	return nil, nil
}
