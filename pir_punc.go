package boosted

import (
	"errors"
	"fmt"
	"io"
	"math"

	//	"log"

	"math/rand"

	//	"sort"

	"github.com/lukechampine/fastxor"
)

type pirClientPunc struct {
	nRows   int
	setSize int

	sets  []PuncturableSet
	hints []Row

	randSource *rand.Rand
	setGen     *shiftedSetGenerator
}

type pirServerPunc struct {
	nRows  int
	rowLen int

	numHintsMultiplier int

	flatDb []byte
}

const Left int = 0
const Right int = 1

func xorInto(a []byte, b []byte) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	fastxor.Bytes(a, a, b)

	// for i := 0; i < len(a); i++ {
	// 	a[i] = a[i] ^ b[i]
	// }
}

func (s *pirServerPunc) xorRowsFlatSlice(out []byte, rows Set) int {
	bytes := 0
	//	setS := setToSlice(set)
	for _, row := range rows {
		if row >= s.nRows {
			continue
		}
		end := s.rowLen*row + len(out)
		if end > len(s.flatDb) {
			end = len(s.flatDb)
		}
		effLen := end - s.rowLen*row
		//fmt.Printf("start: %d, end: %d, effLen: %d, len(s.flatDb): %d", s.rowLen*row, end, effLen, len(s.flatDb))
		xorInto(out[0:effLen], s.flatDb[s.rowLen*row:end])
		bytes += effLen
	}
	return bytes
}

func NewPirServerPunc(source *rand.Rand, data []Row) pirServerPunc {
	s := pirServerPunc{
		nRows:              len(data),
		flatDb:             flattenDb(data),
		numHintsMultiplier: int(float64(*SecParam) * math.Log(2)),
	}
	if len(data) > 0 {
		s.rowLen = len(data[0])
	}
	return s
}

func setToSlice(set Set) []int {
	out := make([]int, len(set))
	i := 0
	for _, k := range set {
		out[i] = k
		i += 1
	}
	return out
}

func (s pirServerPunc) Hint(req HintReq, resp *HintResp) error {
	setSize := int(math.Round(math.Pow(float64(s.nRows), 0.5)))
	nHints := int(math.Round(math.Pow(float64(s.nRows), 0.5))) * s.numHintsMultiplier

	key := make([]byte, 16)
	if _, err := io.ReadFull(rand.New(rand.NewSource(req.RandSeed)), key); err != nil {
		panic(err)
	}

	sets := make([]PuncturableSet, nHints)
	hints := make([]Row, nHints)
	totalRows := 0
	setGen := NewSetGenerator(NewGGMSetGenerator, key)
	for i := 0; i < nHints; i++ {
		var set Set
		sets[i], set = setGen.SetGenAndEval(s.nRows, setSize)
		hints[i] = make(Row, s.rowLen)
		totalRows += len(set)
		xorRowsFlatSlice(s.flatDb, s.rowLen, set, hints[i])
	}
	//fmt.Printf("nHints: %d, total Rows: %d \n", req.NumHints, totalRows)
	resp.Hints = hints
	resp.NumRows = s.nRows
	resp.SetSize = setSize
	resp.SetGenKey = key
	return nil
}

func (s pirServerPunc) dbElem(i int) Row {
	if i < s.nRows {
		return s.flatDb[s.rowLen*i : s.rowLen*(i+1)]
	} else {
		return make(Row, s.rowLen)
	}
}

func (s pirServerPunc) Answer(q QueryReq, resp *QueryResp) error {
	if q.BatchReqs != nil {
		return s.answerBatch(q.BatchReqs, &resp.BatchResps)
	}
	resp.Answer = make(Row, s.rowLen)
	s.xorRowsFlatSlice(resp.Answer, q.PuncturedSet.Eval())
	resp.ExtraElem = s.dbElem(q.ExtraElem)

	// Debug
	resp.Val = s.dbElem(q.Index)

	return nil
}

func (s pirServerPunc) answerBatch(queries []QueryReq, resps *[]QueryResp) error {
	totalRows := 0
	*resps = make([]QueryResp, len(queries))
	for i, q := range queries {
		totalRows += q.PuncturedSet.Size()
		err := s.Answer(q, &(*resps)[i])
		if err != nil {
			return err
		}
	}
	//fmt.Printf("AnswerBatch total rows read: %d\n", totalRows)
	return nil
}

func NewPirClientPunc(source *rand.Rand) *pirClientPunc {
	return &pirClientPunc{randSource: source}
}

func (c *pirClientPunc) initHint(resp *HintResp) error {
	c.nRows = resp.NumRows
	c.setSize = resp.SetSize
	c.hints = resp.Hints
	c.setGen = NewSetGenerator(NewGGMSetGenerator, resp.SetGenKey)

	return nil
}

func (c *pirClientPunc) initSets() {
	c.sets = make([]PuncturableSet, len(c.hints))
	for i := 0; i < len(c.hints); i++ {
		c.sets[i] = c.setGen.SetGen(c.nRows, c.setSize)
	}
	// Use a separate set generator with a new key for all future sets
	// since they must look random to the left server.
	newSetGenKey := make([]byte, 16)
	io.ReadFull(c.randSource, newSetGenKey)
	c.setGen = NewSetGenerator(NewGGMSetGenerator, newSetGenKey)
}

// Sample a biased coin that comes up heads (true) with
// probability (nHeads/total).
func (c *pirClientPunc) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
}

func (c *pirClientPunc) sample(odd1 int, odd2 int, total int) int {
	coin := c.randSource.Intn(total)
	if coin < odd1 {
		return 1
	} else if coin < odd1+odd2 {
		return 2
	} else {
		return 0
	}
}

func (c *pirClientPunc) findIndex(i int) int {
	for j, set := range c.sets {
		if set.Contains(i) {
			return j
		}
	}
	return -1
}

type puncQueryCtx struct {
	i        int
	randCase int
	setIdx   int
}

func (c *pirClientPunc) query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	if len(c.sets) < 1 {
		c.initSets()
	}

	var set PuncturableSet
	ctx := puncQueryCtx{i: i}
	if ctx.setIdx = c.findIndex(i); ctx.setIdx < 0 {
		return nil, nil
	}
	set = c.sets[ctx.setIdx]

	var puncSetL, puncSetR SuccinctSet
	var extraL, extraR int
	ctx.randCase = c.sample(c.setSize-1, c.setSize-1, c.nRows)
	switch ctx.randCase {
	case 0:
		newSet := c.setGen.GenWith(c.nRows, c.setSize, i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(set, i)
		puncSetL = newSet.Punc(i)
		puncSetR = set.Punc(i)
		if ctx.setIdx >= 0 {
			c.sets[ctx.setIdx] = newSet
		}
	case 1:
		newSet := c.setGen.GenWith(c.nRows, c.setSize, i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(newSet, extraL)
		puncSetL = newSet.Punc(extraR)
		puncSetR = newSet.Punc(i)
	case 2:
		newSet := c.setGen.GenWith(c.nRows, c.setSize, i)
		extraR = c.randomMemberExcept(newSet, i)
		extraL = c.randomMemberExcept(newSet, extraR)
		puncSetL = newSet.Punc(i)
		puncSetR = newSet.Punc(extraL)
	}

	return []QueryReq{
			QueryReq{PuncturedSet: puncSetL, ExtraElem: extraL, Index: i /* Debug */},
			QueryReq{PuncturedSet: puncSetR, ExtraElem: extraR, Index: i /* Debug */}},
		func(resp []QueryResp) (Row, error) {
			return c.reconstruct(ctx, resp)
		}
}

func (c *pirClientPunc) dummyQuery() []QueryReq {
	newSetGenKey := make([]byte, 16)
	io.ReadFull(c.randSource, newSetGenKey)
	setGen := NewSetGenerator(NewGGMSetGenerator, newSetGenKey)
	newSet := setGen.GenWith(c.nRows, c.setSize, 0)
	extra := c.randomMemberExcept(newSet, 0)
	puncSet := newSet.Punc(0)
	q := QueryReq{PuncturedSet: puncSet, ExtraElem: extra, Index: 0}
	return []QueryReq{q, q}
}

func (c *pirClientPunc) reconstruct(ctx puncQueryCtx, resp []QueryResp) (Row, error) {
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if ctx.setIdx < 0 {
		return nil, errors.New("couldn't find element in collection")
	}

	switch ctx.randCase {
	case 0:
		hint := c.hints[ctx.setIdx]
		xorInto(out, hint)
		xorInto(out, resp[Right].Answer)
		// Update hint with refresh info
		xorInto(hint, hint)
		xorInto(hint, resp[Left].Answer)
		xorInto(hint, out)
	case 1:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Right].ExtraElem)
	case 2:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Left].ExtraElem)
	}
	return out, nil
}

func (c *pirClientPunc) NumCovered() int {
	covered := make(map[int]bool)
	for _, set := range c.sets {
		for _, elem := range set.Eval() {
			covered[elem] = true
		}
	}
	return len(covered)
}

// Sample a random element of the set that is not equal to `idx`.
func (c *pirClientPunc) randomMemberExcept(set PuncturableSet, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := set.ElemAt(c.randSource.Intn(c.setSize))
		if val != idx {
			return val
		}
	}
}
