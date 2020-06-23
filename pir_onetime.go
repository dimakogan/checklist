package boosted

import (
	//"log"
	"fmt"
	"math"
	"math/rand"
)

type pirClientOneTime struct {
	nRows   int
	rowLen  int
	xorLeft []Row

	randSource *rand.Rand
}

type pirOneTime struct {
	nRows  int
	rowLen int
	flatDb []byte

	randSource *rand.Rand
}

type pirOneTimeHintReq [][]int

type pirOneTimeHintResp []byte

func NewPirServerOneTime(source *rand.Rand, data []Row, hintStrategy int) *pirOneTime {
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

	return &pirOneTime{
		rowLen:     rowLen,
		flatDb:     flatDb,
		randSource: source,
		nRows:      len(data),
	}
}

func (s *pirOneTime) Hint(req pirOneTimeHintReq, resp *pirOneTimeHintResp) error {
	hint := make([]byte, len(req)*s.rowLen)

	bytes := 0
	for i := 0; i < len(req); i++ {
		for j := 0; j < len(req[i]); j++ {
			xorInto(hint[i*s.rowLen:(i+1)*s.rowLen],
				s.flatDb[s.rowLen*req[i][j]:(s.rowLen*(req[i][j]+1))])
			bytes += s.rowLen
		}
	}
	//log.Printf("Bytes: %v", bytes)

	*resp = hint
	return nil
}

func (s *pirOneTime) Answer(q *QueryReq, resp *QueryResp) error {
	return nil
}

func newPirClientOneTime(source *rand.Rand, nRows int, rowLen int) *pirClientOneTime {
	return &pirClientOneTime{
		rowLen:     rowLen,
		randSource: source,
		nRows:      nRows,
	}
}

func (c *pirClientOneTime) RequestHint() (*pirOneTimeHintReq, error) {
	idx := make([]int, c.nRows)
	for i := 0; i < c.rowLen; i++ {
		idx[i] = i
	}

	c.randSource.Shuffle(c.rowLen, func(i int, j int) {
		t := idx[i]
		idx[i] = idx[j]
		idx[j] = t
	})

	nSets := int(math.Sqrt(float64(c.nRows)))
	var hr pirOneTimeHintReq
	hr = make([][]int, nSets)
	setSize := c.nRows / nSets
	for i := 0; i < nSets; i++ {
		hr[i] = make([]int, setSize)

		if setSize*(i+1) < len(idx) {
			copy(hr[i][:], idx[setSize*i:setSize*(i+1)])
		} else {
			copy(hr[i][:], idx[setSize*i:])
		}
	}

	return &hr, nil
}

func (c *pirClientOneTime) InitHint(resp *HintResp) error {
	return nil
}

func (c *pirClientOneTime) Query(i int) ([]*QueryReq, error) {
	return []*QueryReq{}, nil
}

func (c *pirClientOneTime) Reconstruct(resp []*QueryResp) (Row, error) {
	if len(resp) != 1 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(resp))
	}

	return nil, nil
}
