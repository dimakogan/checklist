package boosted

import (
	"errors"
	"fmt"
//	"log"
	"math"
	"math/rand"
//	"sort"

	"github.com/lukechampine/fastxor"
)

type HintFunc func(s *pirServerPunc, req *HintReq, resp *HintResp) error

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

	flatDb []byte

	hintFunc   HintFunc
	randSource *rand.Rand
}

func xorInto(a []byte, b []byte) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	fastxor.Bytes(a, a, b)

	// for i := 0; i < len(a); i++ {
	// 	a[i] = a[i] ^ b[i]
	// }
}

func (s *pirServerPunc) xorRows(out Row, rows Set, delta int) {
	// TODO: Parallelize this function.
	for row := range rows {
		xorInto(out, s.db[(row+delta)%len(s.db)])
	}
}

func (s *pirServerPunc) xorRowsFlatSlice(out Row, rows []int, delta int) int {
  bytes := 0
	for _, row := range rows {
		drow := (row + delta) % len(s.db)
		xorInto(out, s.flatDb[s.rowLen*drow:s.rowLen*(drow+1)])
    bytes += s.rowLen
	}
  return bytes
}

func NewPirServerPunc(source *rand.Rand, data []Row) PIRServer {
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

	return &pirServerPunc{
		rowLen:     rowLen,
		db:         data,
		flatDb:     flatDb,
		randSource: source,
	}
}

func setToSlice(set Set) []int {
	out := make([]int, len(set))
	i := 0
	for k := range set {
		out[i] = k
		i += 1
	}
	return out
}


func (s *pirServerPunc) Hint(req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)

	set := req.Key.Eval()
  setS := setToSlice(set)

  bytes := 0
	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)
    bytes = bytes + s.xorRowsFlatSlice(hints[j], setS, req.Deltas[j])
	}
  //log.Printf("bytes: %v", bytes)

	resp.Hints = hints
	return nil
}

func (s *pirServerPunc) Answer(q *QueryReq, resp *QueryResp) error {
	rows := q.Key.Eval()
	resp.Answer = make(Row, s.rowLen)
	s.xorRows(resp.Answer, rows, 0)
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
  return c.RequestHintN(nHints)
}

func (c *pirClientPunc) RequestHintN(nHints int) (*HintReq, error) {
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

func (c *pirClientPunc) CanQuery(i int) bool {
	return c.key.FindShift(i, c.deltas) >= 0
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
