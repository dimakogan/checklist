package pir

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"

	"math/rand"

	"checklist/psetggm"
)

type puncClient struct {
	nRows  int
	RowLen int

	setSize int

	sets []SetKey

	hints []Row

	randSource         *rand.Rand
	origSetGen, setGen SetGenerator

	idxToSetIdx []int32
}

type PuncHintReq struct {
	RandSeed           PRGKey
	NumHintsMultiplier int
}

type PuncHintResp struct {
	NRows     int
	RowLen    int
	SetSize   int
	SetGenKey PRGKey
	Hints     []Row
}

func xorRowsFlatSlice(db *StaticDB, out []byte, indices Set) {
	for i := range indices {
		indices[i] *= db.RowLen
	}
	psetggm.XorBlocks(db.FlatDb, indices, out)

}

func NewPuncHintReq(randSource *rand.Rand) *PuncHintReq {
	req := &PuncHintReq{
		RandSeed:           PRGKey{},
		NumHintsMultiplier: int(float64(SecParam) * math.Log(2)),
	}
	_, err := io.ReadFull(randSource, req.RandSeed[:])
	if err != nil {
		log.Fatalf("Failed to initialize random seed: %s", err)
	}
	return req
}

func (req *PuncHintReq) Process(db StaticDB) (HintResp, error) {
	setSize := int(math.Round(math.Pow(float64(db.NumRows), 0.5)))
	nHints := req.NumHintsMultiplier * db.NumRows / setSize

	hints := make([]Row, nHints)
	hintBuf := make([]byte, db.RowLen*nHints)
	setGen := NewSetGenerator(req.RandSeed, 0, db.NumRows, setSize)
	var pset PuncturableSet
	for i := 0; i < nHints; i++ {
		setGen.Gen(&pset)
		hints[i] = Row(hintBuf[db.RowLen*i : db.RowLen*(i+1)])
		xorRowsFlatSlice(&db, hints[i], pset.elems)
	}

	return &PuncHintResp{
		Hints:     hints,
		NRows:     db.NumRows,
		RowLen:    db.RowLen,
		SetSize:   setSize,
		SetGenKey: req.RandSeed,
	}, nil
}

func dbElem(db StaticDB, i int) Row {
	if i < db.NumRows {
		return db.Row(i)
	} else {
		return make(Row, db.RowLen)
	}
}

func (resp *PuncHintResp) InitClient(source *rand.Rand) Client {
	c := puncClient{
		randSource: source,
		nRows:      resp.NRows,
		RowLen:     resp.RowLen,
		setSize:    resp.SetSize,
		hints:      resp.Hints,
		origSetGen: NewSetGenerator(resp.SetGenKey, 0, resp.NRows, resp.SetSize),
	}
	c.initSets()
	return &c
}

func (resp *PuncHintResp) NumRows() int {
	return resp.NRows
}

func (c *puncClient) initSets() {
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
	var newSetGenKey PRGKey
	io.ReadFull(c.randSource, newSetGenKey[:])
	c.setGen = NewSetGenerator(newSetGenKey, c.origSetGen.num, c.nRows, c.setSize)
}

// Sample a biased coin that comes up heads (true) with
// probability (nHeads/total).
func (c *puncClient) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
}

func (c *puncClient) sample(odd1 int, odd2 int, total int) int {
	coin := c.randSource.Intn(total)
	if coin < odd1 {
		return 1
	} else if coin < odd1+odd2 {
		return 2
	} else {
		return 0
	}
}

func (c *puncClient) findIndex(i int) (setIdx int) {
	if i >= c.nRows {
		return -1
	}

	if setIdx := c.idxToSetIdx[MathMod(i, c.nRows)]; setIdx >= 0 {
		return int(setIdx)
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
			if shiftedV == i {
				return j
			}

			if shiftedV < c.nRows {
				// upgrade invalid pointer to valid one
				c.idxToSetIdx[shiftedV] = int32(j)
			}
		}
	}
	return -1
}

type PuncQueryReq struct {
	PuncturedSet PuncturedSet
	ExtraElem    int
}

type PuncQueryResp struct {
	Answer    Row
	ExtraElem Row
}

type puncQueryCtx struct {
	randCase int
	setIdx   int
}

func (c *puncClient) Query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}

	var ctx puncQueryCtx

	if ctx.setIdx = c.findIndex(i); ctx.setIdx < 0 {
		return nil, nil
	}
	i = MathMod(i, c.nRows)

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
			&PuncQueryReq{PuncturedSet: puncSetL, ExtraElem: extraL},
			&PuncQueryReq{PuncturedSet: puncSetR, ExtraElem: extraR},
		},
		func(resps []interface{}) (Row, error) {
			queryResps := make([]*PuncQueryResp, len(resps))
			var ok bool
			for i, r := range resps {
				if queryResps[i], ok = r.(*PuncQueryResp); !ok {
					return nil, fmt.Errorf("Invalid response type: %T, expected: *PuncQueryResp", r)
				}
			}

			return c.reconstruct(ctx, queryResps)
		}
}

func (c *puncClient) eval(setIdx int) PuncturableSet {
	if c.sets[setIdx].id < c.origSetGen.num {
		return c.origSetGen.Eval(c.sets[setIdx])
	} else {
		return c.setGen.Eval(c.sets[setIdx])
	}
}

func (c *puncClient) setGenForSet(setIdx int) *SetGenerator {
	if c.sets[setIdx].id < c.origSetGen.num {
		return &c.origSetGen
	} else {
		return &c.setGen
	}
}

func (c *puncClient) replaceSet(setIdx int, newSet PuncturableSet) {
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

func (c *puncClient) DummyQuery() []QueryReq {
	newSet := c.setGen.GenWith(0)
	extra := c.randomMemberExcept(newSet, 0)
	puncSet := c.setGen.Punc(newSet, 0)
	q := PuncQueryReq{PuncturedSet: puncSet, ExtraElem: extra}
	return []QueryReq{&q, &q}
}

func (q *PuncQueryReq) Process(db StaticDB) (interface{}, error) {
	resp := PuncQueryResp{Answer: make(Row, db.RowLen)}
	psetggm.FastAnswer(q.PuncturedSet.Keys, q.PuncturedSet.Hole, q.PuncturedSet.UnivSize, q.PuncturedSet.SetSize, int(q.PuncturedSet.Shift),
		db.FlatDb, db.RowLen, resp.Answer)
	resp.ExtraElem = db.FlatDb[db.RowLen*q.ExtraElem : db.RowLen*q.ExtraElem+db.RowLen]

	return &resp, nil
}

func (c *puncClient) reconstruct(ctx puncQueryCtx, resp []*PuncQueryResp) (Row, error) {
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

func (c *puncClient) NumCovered() int {
	covered := make(map[int]bool)
	for j := range c.sets {
		for _, elem := range c.eval(j).elems {
			covered[elem] = true
		}
	}
	return len(covered)
}

// Sample a random element of the set that is not equal to `idx`.
func (c *puncClient) randomMemberExcept(set PuncturableSet, idx int) int {
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

func (c *puncClient) StateSize() (bitsPerKey, fixedBytes int) {
	return int(math.Log2(float64(len(c.hints)))), len(c.hints) * c.RowLen
}
