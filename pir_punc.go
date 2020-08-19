package boosted

import (
	"errors"
	"fmt"
	"io"
	"sync"

	//	"log"
	"math"
	"math/rand"

	//	"sort"

	"github.com/lukechampine/fastxor"
)

type pirClientPunc struct {
	nRows   int
	nHints  int
	setSize int

	sets  []PuncturableSet
	hints []Row

	randSource *rand.Rand
	setGen     *shiftedSetGenerator

	servers [2]PirServer
}

type pirServerPunc struct {
	nRows  int
	rowLen int

	flatDb []byte
}

const Left int = 0
const Right int = 1

type PirServer interface {
	Hint(req *HintReq, resp *HintResp) error
	AnswerBatch(q []QueryReq, resp *[]QueryResp) error
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
	return pirServerPunc{
		rowLen: len(data[0]),
		nRows:  len(data),
		flatDb: flattenDb(data)}
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

func genSets(masterKey []byte, nSets, setSize, univSize int) []PuncturableSet {
	setGen := NewSetGenerator(NewGGMSetGenerator, masterKey)
	sets := make([]PuncturableSet, nSets)
	for i := 0; i < nSets; i++ {
		sets[i] = setGen.SetGen(univSize, setSize)
	}
	return sets
}

func (s pirServerPunc) Hint(req *HintReq, resp *HintResp) error {
	sets := genSets(req.SetGenKey, req.NumHints, req.SetSize, s.nRows)
	hints := make([]Row, req.NumHints)

	totalRows := 0

	for j := 0; j < req.NumHints; j++ {
		hints[j] = make(Row, s.rowLen)
		set := sets[j].Eval()
		totalRows += len(set)
		xorRowsFlatSlice(s.flatDb, s.rowLen, set, hints[j])
	}
	//fmt.Printf("nHints: %d, total Rows: %d \n", req.NumHints, totalRows)
	resp.Hints = hints
	return nil
}

func (s pirServerPunc) dbElem(i int) Row {
	if i < s.nRows {
		return s.flatDb[s.rowLen*i : s.rowLen*(i+1)]
	} else {
		return make(Row, s.rowLen)
	}
}

func (s pirServerPunc) answer(q QueryReq, resp *QueryResp) error {
	resp.Answer = make(Row, s.rowLen)
	s.xorRowsFlatSlice(resp.Answer, q.PuncturedSet.Eval())
	resp.ExtraElem = s.dbElem(q.ExtraElem)

	// Debug
	resp.Val = s.dbElem(q.Index)
	return nil
}

func (s pirServerPunc) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	totalRows := 0
	*resps = make([]QueryResp, len(queries))
	for i, q := range queries {
		totalRows += q.PuncturedSet.Size()
		err := s.answer(q, &(*resps)[i])
		if err != nil {
			return err
		}
	}
	//fmt.Printf("AnswerBatch total rows read: %d\n", totalRows)
	return nil
}

func NewPirClientPunc(source *rand.Rand, nRows int, servers [2]PirServer) *pirClientPunc {
	// TODO: Maybe better to just do this with integer ops.
	// nRowsRounded := 1 << int(math.Ceil(math.Log2(float64(nRows))/2)*2)
	nRowsRounded := nRows
	setSize := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))
	nHints := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))

	return &pirClientPunc{
		nRows:      nRowsRounded,
		setSize:    setSize,
		nHints:     nHints,
		hints:      nil,
		randSource: source,
		servers:    servers,
	}
}

func (c *pirClientPunc) requestHint() (*HintReq, error) {
	setGenKey := make([]byte, 16)
	io.ReadFull(c.randSource, setGenKey)
	c.sets = genSets(setGenKey, c.nHints, c.setSize, c.nRows)

	return &HintReq{
		NumHints:  c.nHints,
		SetSize:   c.setSize,
		SetGenKey: setGenKey}, nil
}

func (c *pirClientPunc) initHint(resp *HintResp) error {
	c.hints = resp.Hints
	// Use a separate set generator with a new key for all future sets
	// since they must look random to the left server.
	newSetGenKey := make([]byte, 16)
	io.ReadFull(c.randSource, newSetGenKey)
	c.setGen = NewSetGenerator(NewGGMSetGenerator, newSetGenKey)

	return nil
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

func (c *pirClientPunc) query(i int) ([]QueryReq, puncQueryCtx) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}

	var set PuncturableSet
	ctx := puncQueryCtx{i: i}
	if ctx.setIdx = c.findIndex(i); ctx.setIdx >= 0 {
		set = c.sets[ctx.setIdx]
	} else {
		set = c.setGen.GenWith(c.nRows, c.setSize, i)
	}

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
		ctx
}

func (c *pirClientPunc) reconstruct(ctx puncQueryCtx, resp []*QueryResp) (Row, error) {
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

func (c *pirClientPunc) Init() error {
	hintReq, err := c.requestHint()
	if err != nil {
		return err
	}
	var hintResp HintResp
	err = c.servers[Left].Hint(hintReq, &hintResp)
	if err != nil {
		return err
	}
	return c.initHint(&hintResp)
}

func (c *pirClientPunc) ReadBatch(idxs []int) ([]Row, []error) {
	return c.ReadBatchAtLeast(idxs, len(idxs))
}

func (c *pirClientPunc) ReadBatchAtLeast(idxs []int, n int) ([]Row, []error) {
	reqs := [][]QueryReq{make([]QueryReq, n), make([]QueryReq, n)}
	ctxs := make([]puncQueryCtx, len(idxs))

	nOk := 0
	for pos, i := range idxs {
		var queryReqs []QueryReq
		queryReqs, ctxs[pos] = c.query(i)
		if ctxs[pos].setIdx < 0 {
			continue
		}
		reqs[Left][nOk] = queryReqs[Left]
		reqs[Right][nOk] = queryReqs[Right]
		nOk++
		// Only issue first n non-failing queries
		if nOk == n {
			break
		}
	}

	resps := [][]QueryResp{make([]QueryResp, nOk), make([]QueryResp, nOk)}
	err := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for side := range []int{Left, Right} {
		go func(side int) {
			err[side] = c.servers[side].AnswerBatch(reqs[side][0:nOk], &resps[side])
			wg.Done()
		}(side)
	}
	wg.Wait()

	errs := make([]error, len(idxs))
	if err[Left] != nil {
		err[Right] = err[Left]
	}
	if err[Right] != nil {
		for i := 0; i < len(idxs); i++ {
			errs[i] = err[Right]
		}
		return nil, errs
	}
	vals := make([]Row, len(idxs))
	nOk = 0
	for i := 0; i < len(idxs); i++ {
		if ctxs[i].setIdx >= 0 && nOk < n {
			vals[i], errs[i] = c.reconstruct(ctxs[i], []*QueryResp{&resps[Left][nOk], &resps[Right][nOk]})
			nOk++
		} else {
			vals[i] = nil
			errs[i] = errors.New("couldn't find element in collection")
		}
	}
	return vals, errs
}

func (c *pirClientPunc) Read(i int) (Row, error) {
	vals, err := c.ReadBatch([]int{i})
	return vals[0], err[0]
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
