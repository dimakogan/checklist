package boosted

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

type pirClientPunc struct {
	nRows   int
	setSize int

	key      *SetKey
	deltas   []int
	shiftIdx int
	hints    []Row

	randSource *rand.Rand
}

type pirServerPunc struct {
	rowLen int
	db     []Row

	randSource *rand.Rand
}

func xorInto(a Row, b Row) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	for i := 0; i < len(a); i++ {
		a[i] = a[i] ^ b[i]
	}
}

func (s *pirServerPunc) xorRows(out Row, rows Set) {
	// TODO: Parallelize this function.
	for row := range rows {
		xorInto(out, s.db[row])
	}
}

func newPirServerPunc(source *rand.Rand, data []Row) PIRServer {
	if len(data) < 1 {
		panic("Database must contain at least one row")
	}

	rowLen := len(data[0])
	for _, v := range data {
		if len(v) != rowLen {
			panic("Database rows must all be of the same length")
		}
	}

	return &pirServerPunc{
		rowLen:     rowLen,
		db:         data,
		randSource: source,
	}
}

func (s *pirServerPunc) Hint(req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)

	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)
		req.Key.Shift(req.Deltas[j])
		set := req.Key.Eval()
		s.xorRows(hints[j], set)
		req.Key.Shift(-req.Deltas[j])
	}

	resp.Hints = hints
	return nil
}

func (s *pirServerPunc) Answer(q *QueryReq, resp *QueryResp) error {
	rows := q.Key.Eval()
	resp.Answer = make(Row, s.rowLen)
	s.xorRows(resp.Answer, rows)
	return nil
}

func newPirClientPunc(source *rand.Rand, nRows int) PIRClient {
	// TODO: Maybe better to just do this with integer ops.
	nf := float64(nRows)
	setSize := int(math.Round(math.Pow(nf, 0.5)))

	return &pirClientPunc{
		nRows:      nRows,
		setSize:    setSize,
		hints:      nil,
		randSource: source,
	}
}

func (c *pirClientPunc) RequestHint() (*HintReq, error) {
	nHints := c.setSize * int(math.Round(math.Log2(float64(c.nRows))))
	c.deltas = make([]int, nHints)
	for i := range c.deltas {
		c.deltas[i] = c.randSource.Intn(c.nRows)
	}

	c.key = SetGen(c.randSource, c.nRows, c.setSize)
	return &HintReq{
		Key:    c.key,
		Deltas: c.deltas,
	}, nil
}

func (c *pirClientPunc) InitHint(resp *HintResp) error {
	c.hints = resp.Hints
	return nil
}

// Sample a biased coin that comes up heads (true) with
// probability (nHeads/total).
func (c *pirClientPunc) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
}

func (c *pirClientPunc) Query(i int) ([]*QueryReq, error) {
	if len(c.hints) < 1 {
		return nil, fmt.Errorf("No stored hints. Did you forget to call InitHint?")
	}

	c.shiftIdx = c.key.FindShift(i, c.deltas)

	if c.shiftIdx >= 0 {
		c.key.Shift(c.deltas[c.shiftIdx])
	} else {
		iPrime := c.key.RandomMember(c.randSource)
		shift := MathMod(i-iPrime, c.nRows)
		c.key.Shift(shift)
	}

	coin := c.bernoulli(c.setSize-1, c.nRows)
	var iPunc int
	if coin {
		iPunc = c.key.RandomMemberExcept(c.randSource, i)
		c.shiftIdx = -1
	} else {
		iPunc = i
	}

	return []*QueryReq{
		&QueryReq{Key: c.key.Punc(iPunc)},
	}, nil
}

func (c *pirClientPunc) Reconstruct(resp []*QueryResp) (Row, error) {
	if len(resp) != 1 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if c.shiftIdx < 0 {
		return nil, errors.New("Fail")
	} else {
		xorInto(out, c.hints[c.shiftIdx])
		xorInto(out, resp[0].Answer)
	}

	return out, nil
}
