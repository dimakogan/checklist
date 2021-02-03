package boosted

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"

	//	"log"

	"math/rand"

	//	"sort"

	"github.com/dimakogan/boosted-pir/psetggm"
	"github.com/lukechampine/fastxor"
)

var nRowsPerBlock *int = flag.Int("numRowsPerBlock", 1, "Num DB Rows in each block")

type pirClientPunc struct {
	nRows  int
	rowLen int

	setSize int

	sets []SetKey

	nRowsPerBlock int
	hints         []Row

	randSource         *rand.Rand
	origSetGen, setGen SetGenerator

	idxToSetIdx []int32
}

type pirServerPunc struct {
	nRows  int
	rowLen int

	numHintsMultiplier int
	nRowsPerBlock      int
	blockLen           int

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

func (s *pirServerPunc) xorRowsFlatSlice(out []byte, indices Set) {
	xorRowsFlatSlice(s.flatDb, s.rowLen, indices, out)
}

func NewPirServerPunc(source *rand.Rand, flatDb []byte, nRows, rowLen int) pirServerPunc {
	s := pirServerPunc{
		flatDb:             flatDb,
		nRows:              nRows,
		rowLen:             rowLen,
		numHintsMultiplier: int(float64(*SecParam) * math.Log(2)),
		nRowsPerBlock:      *nRowsPerBlock,
	}

	s.blockLen = s.rowLen * (s.nRowsPerBlock)

	// Make flat DB look cyclical
	//	copy(s.flatDb[s.nRows*s.rowLen:], s.flatDb[0:s.blockLen])

	return s
}

func (s pirServerPunc) Hint(req HintReq, resp *HintResp) error {
	setSize := int(math.Round(math.Pow(float64(s.nRows), 0.5)))
	nHints := s.numHintsMultiplier * s.nRows / (setSize * s.nRowsPerBlock)

	key := make([]byte, 16)
	if _, err := io.ReadFull(rand.New(rand.NewSource(req.RandSeed)), key); err != nil {
		panic(err)
	}

	hints := make([]Row, nHints)
	hintBuf := make([]byte, s.blockLen*nHints)
	setGen := NewSetGenerator(key, 0, s.nRows, setSize)
	var pset PuncturableSet
	for i := 0; i < nHints; i++ {
		setGen.Gen(&pset)
		hints[i] = Row(hintBuf[s.blockLen*i : s.blockLen*(i+1)])
		s.xorRowsFlatSlice(hints[i], pset.elems)
	}
	resp.Hints = hints
	resp.NumRows = s.nRows
	resp.NumRowsPerBlock = s.nRowsPerBlock
	resp.RowLen = s.rowLen
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

func (s pirServerPunc) NumRows(none int, out *int) error {
	*out = s.nRows
	return nil
}

func (s pirServerPunc) GetRow(idx int, row *RowIndexVal) error {
	if idx == -1 {
		// return random row
		idx = RandSource().Int() % s.nRows
	}

	if idx < 0 || idx >= s.nRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.nRows)
	}

	row.Index = idx
	row.Key = uint32(idx)
	row.Value = s.dbElem(idx)
	return nil
}

func (s pirServerPunc) Answer(q QueryReq, resp *QueryResp) error {
	if q.BatchReqs != nil {
		return s.answerBatch(q.BatchReqs, &resp.BatchResps)
	}
	resp.Answer = make(Row, s.blockLen)
	//s.xorRowsFlatSlice(resp.Answer, q.PuncturedSet.Eval())
	psetggm.FastAnswer(q.PuncturedSet.Keys, q.PuncturedSet.Hole, q.PuncturedSet.UnivSize, q.PuncturedSet.SetSize, int(q.PuncturedSet.Shift),
		s.flatDb, s.rowLen, resp.Answer)
	resp.ExtraElem = s.flatDb[s.rowLen*q.ExtraElem : s.rowLen*q.ExtraElem+s.blockLen]

	// Debug
	resp.Val = s.dbElem(q.Index)

	return nil
}

func (s pirServerPunc) answerBatch(queries []QueryReq, resps *[]QueryResp) error {
	totalRows := 0
	*resps = make([]QueryResp, len(queries))
	for i, q := range queries {
		totalRows += q.PuncturedSet.SetSize
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
	c.nRowsPerBlock = resp.NumRowsPerBlock
	c.rowLen = resp.RowLen
	c.setSize = resp.SetSize
	c.hints = resp.Hints
	c.origSetGen = NewSetGenerator(resp.SetGenKey, 0, c.nRows, c.setSize)
	c.initSets()
	return nil
}

func (c *pirClientPunc) initSets() {
	c.sets = make([]SetKey, len(c.hints))
	c.idxToSetIdx = make([]int32, c.nRows)
	for i := range c.idxToSetIdx {
		c.idxToSetIdx[i] = -1
	}
	var pset PuncturableSet
	for i := 0; i < len(c.hints); i++ {
		c.origSetGen.Gen(&pset)
		c.sets[i] = pset.SetKey
		for _, j := range pset.elems {
			c.idxToSetIdx[j] = int32(i)
		}
	}

	// Use a separate set generator with a new key for all future sets
	// since they must look random to the left server.
	newSetGenKey := make([]byte, 16)
	io.ReadFull(c.randSource, newSetGenKey)
	c.setGen = NewSetGenerator(newSetGenKey, c.origSetGen.num, c.nRows, c.setSize)
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


func (c *pirClientPunc) findIndex(i int) (setIdx, posInBlock int) {
	if i >= c.nRows {
		return -1, -1
	}

	for posInBlock = 0; posInBlock < c.nRowsPerBlock; posInBlock++ {
		if setIdx := c.idxToSetIdx[MathMod(i-posInBlock, c.nRows)]; setIdx >= 0 {
			return int(setIdx), posInBlock
		}
	}
	var pset PuncturableSet
	// If set pointer of i is invalid, use this opportunity to upgrade other invalid pointers while doing linear scan
	for j := range c.sets {
		setGen := c.setGenForSet(j)
		setKeyNoShift := c.sets[j]
		shift := setKeyNoShift.shift
		setKeyNoShift.shift = 0
		setGen.EvalInPlace(setKeyNoShift, &pset)

		for _, v := range pset.elems {
			shiftedV := int((uint32(v) + shift) % uint32(setGen.univSize))
			for posInBlock = 0; posInBlock < c.nRowsPerBlock; posInBlock++ {
				if shiftedV+posInBlock == i {
					return j, posInBlock
				}
			}
			if shiftedV < c.nRows {
				// upgrade invalid pointer to valid one
				c.idxToSetIdx[shiftedV] = int32(j)
			}
		}
	}
	return -1, -1
}

type puncQueryCtx struct {
	posInBlock int
	randCase   int
	setIdx     int
}

func (c *pirClientPunc) query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}

	var ctx puncQueryCtx

	if ctx.setIdx, ctx.posInBlock = c.findIndex(i); ctx.setIdx < 0 {
		return nil, nil
	}
	origI := i
	i = MathMod(i-ctx.posInBlock, c.nRows)

	pset := c.eval(ctx.setIdx)

	var puncSetL, puncSetR PuncturedSet
	var extraL, extraR int
	ctx.randCase = c.sample(c.setSize-1, c.setSize-1, c.nRows)
	switch ctx.randCase {
	case 0:
		newSet := c.setGen.GenWith(i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(pset, i)
		puncSetL = c.setGen.Punc(newSet, i)
		puncSetR = c.setGen.Punc(pset, i)
		if ctx.setIdx >= 0 {
			c.replaceSet(ctx.setIdx, newSet)
		}
	case 1:
		newSet := c.setGen.GenWith(i)
		extraR = c.randomMemberExcept(newSet, i)
		extraL = c.randomMemberExcept(newSet, extraR)
		puncSetL = c.setGen.Punc(newSet, extraR)
		puncSetR = c.setGen.Punc(newSet, i)
	case 2:
		newSet := c.setGen.GenWith(i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(newSet, extraL)
		puncSetL = c.setGen.Punc(newSet, i)
		puncSetR = c.setGen.Punc(newSet, extraL)
	}

	return []QueryReq{
			{PuncturedSet: puncSetL, ExtraElem: extraL, Index: origI /* Debug */},
			{PuncturedSet: puncSetR, ExtraElem: extraR, Index: origI /* Debug */}},
		func(resp []QueryResp) (Row, error) {
			return c.reconstruct(ctx, resp)
		}
}

func (c *pirClientPunc) eval(setIdx int) PuncturableSet {
	if c.sets[setIdx].id < c.origSetGen.num {
		return c.origSetGen.Eval(c.sets[setIdx])
	} else {
		return c.setGen.Eval(c.sets[setIdx])
	}
}

func (c *pirClientPunc) setGenForSet(setIdx int) *SetGenerator {
	if c.sets[setIdx].id < c.origSetGen.num {
		return &c.origSetGen
	} else {
		return &c.setGen
	}
}

func (c *pirClientPunc) replaceSet(setIdx int, newSet PuncturableSet) {
	pset := c.eval(setIdx)
	for _, idx := range pset.elems {
		if idx < c.nRows && c.idxToSetIdx[idx] == int32(setIdx) {
			c.idxToSetIdx[idx] = -1
		}
	}

	c.sets[setIdx] = newSet.SetKey
	for _, v := range newSet.elems {
		c.idxToSetIdx[v] = int32(setIdx)
	}
}

func (c *pirClientPunc) dummyQuery() []QueryReq {
	newSet := c.setGen.GenWith(0)
	extra := c.randomMemberExcept(newSet, 0)
	puncSet := c.setGen.Punc(newSet, 0)
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
	return out[ctx.posInBlock*c.rowLen : (ctx.posInBlock+1)*c.rowLen], nil
}

func (c *pirClientPunc) NumCovered() int {
	covered := make(map[int]bool)
	for j := range c.sets {
		for _, elem := range c.eval(j).elems {
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
		val := set.elems[c.randSource.Intn(c.setSize)]
		if val != idx {
			return val
		}
	}
}
