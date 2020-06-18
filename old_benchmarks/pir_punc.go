//+build ignore

package boosted

import (
	"errors"
	"fmt"
	"log"
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

func (s *pirServerPunc) xorRowsFlat(out Row, rows Set, delta int) {
	for row := range rows {
		drow := (row + delta) % len(s.db)
		xorInto(out, s.flatDb[s.rowLen*drow:s.rowLen*(drow+1)])
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

func NewPirServerPunc(source *rand.Rand, data []Row, hintStrategy int) PIRServer {
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

	var hf HintFunc
	switch hintStrategy {
	case 0:
		hf = HintRandom
	case 1:
		hf = HintLinear
		/*
			case 2:
				hf = HintLinearSort
		*/
	case 3:
		hf = HintFlat
	case 4:
		hf = HintFlatLinear
	case 5:
		hf = HintFlatSlice
		/*
			case 6:
				hf = HintFake
		*/
	default:
		panic("Unknown hint type")
	}

	return &pirServerPunc{
		rowLen:     rowLen,
		db:         data,
		hintFunc:   hf,
		flatDb:     flatDb,
		randSource: source,
	}
}

func (s *pirServerPunc) Hint(req *HintReq, resp *HintResp) error {
	return s.hintFunc(s, req, resp)
}

/*
func HintLinearSort(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)
	rowSets := make([][]int, len(s.db))

	for i := 0; i < len(s.db); i++ {
		rowSets[i] = make([]int, 0, 10*nHints)
	}

	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)

		req.Key.Shift(req.Deltas[j])
		set := req.Key.Eval()
		for k := range set {
			rowSets[k] = append(rowSets[k], j)
		}

		req.Key.Shift(-req.Deltas[j])
	}

	for i := 0; i < len(s.db); i++ {
		row := s.db[i]

		sort.Ints(rowSets[i])
		for _, j := range rowSets[i] {
			xorInto(hints[j], row)
		}
	}

	resp.Hints = hints
	return nil
}
*/

func HintLinear(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)
	rowSets := make([]Set, len(s.db))

	for i := 0; i < len(s.db); i++ {
		rowSets[i] = make(Set)
	}

	set := req.Key.Eval()
	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)

		for k := range set {
			rowSets[(k+req.Deltas[j])%len(s.db)][j] = Present_Yes
		}
	}

	for i := 0; i < len(s.db); i++ {
		row := s.db[i]
		for j := range rowSets[i] {
			xorInto(hints[j], row)
		}
	}

	resp.Hints = hints
	return nil
}

func HintFlatLinear(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)
	rowSets := make([]Set, len(s.db))

	for i := 0; i < len(s.db); i++ {
		rowSets[i] = make(Set)
	}

	set := req.Key.Eval()
	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)

		for k := range set {
			rowSets[(k+req.Deltas[j])%len(s.db)][j] = Present_Yes
		}
	}

	for i := 0; i < len(s.db); i++ {
		row := s.flatDb[i*s.rowLen : (i+1)*s.rowLen]
		for j := range rowSets[i] {
			xorInto(hints[j], row)
		}
	}

	resp.Hints = hints
	return nil
}

func HintRandom(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	return HintRandomType(s, req, resp, false, false)
}

func HintFlat(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	return HintRandomType(s, req, resp, true, false)
}

func HintFlatSlice(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	return HintRandomType(s, req, resp, true, true)
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

/*
func HintFake(s *pirServerPunc, req *HintReq, resp *HintResp) error {
	nHints := len(req.Deltas)
	hints := make([]byte, s.rowLen * nHints)

  bound := int(math.Sqrt(float64(len(s.db))))
  for i := 0; i<len(s.db); i++ {
    row := make(Row, s.rowLen)
    copy(row[:], s.flatDb[i*s.rowLen:(i+1)*s.rowLen])
    for j := 0; j<nHints; j++ {
      if(s.randSource.Intn(len(s.db)) < bound) {
        xorInto(hints[j*s.rowLen:(j+1)*s.rowLen], row[:])
      }
    }
  }

	resp.Hints = make([]Row, nHints)
  for i := 0; i<nHints; i++ {
    resp.Hints[i] = make([]byte, s.rowLen)
    copy(resp.Hints[i][:], hints[i*s.rowLen:(i+1)*s.rowLen])
  }
	return nil
}
*/

func HintRandomType(s *pirServerPunc, req *HintReq, resp *HintResp, flat bool, setSlice bool) error {
	nHints := len(req.Deltas)
	hints := make([]Row, nHints)

	set := req.Key.Eval()
	var setS []int
	if setSlice {
		setS = setToSlice(set)
	}

	bytes := 0
	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)

		if flat {
			if setSlice {
				bytes = bytes + s.xorRowsFlatSlice(hints[j], setS, req.Deltas[j])
			} else {
				s.xorRowsFlat(hints[j], set, req.Deltas[j])
			}
		} else {
			s.xorRows(hints[j], set, req.Deltas[j])
		}
	}
	log.Printf("bytes: %v", bytes)

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
