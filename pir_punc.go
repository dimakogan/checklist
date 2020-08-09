package boosted

import (
	"errors"
	"fmt"
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

	sets       []PuncturableSet
	hints      []Row
	auxRecords map[int]Row

	randSource *rand.Rand
	setGen     SetGenerator

	servers [2]PirServer
}

type pirServerPunc struct {
	nRows  int
	rowLen int

	flatDb []byte

	randSource *rand.Rand
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
		rowLen:     len(data[0]),
		nRows:      len(data),
		flatDb:     flattenDb(data),
		randSource: source,
	}
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

func (s pirServerPunc) Hint(req *HintReq, resp *HintResp) error {
	nHints := len(req.Sets)
	hints := make([]Row, nHints)

	totalRows := 0

	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)
		set := req.Sets[j].Eval()
		totalRows += len(set)
		xorRowsFlatSlice(s.flatDb, s.rowLen, set, hints[j])
	}
	//fmt.Printf("nHints: %d, total Rows: %d \n", nHints, totalRows)
	resp.Hints = hints

	// auxSet := req.AuxRecordsSet.Eval()
	// resp.AuxRecords = make(map[int]Row)
	// for row := range auxSet {
	// 	if row < s.nRows {
	// 		resp.AuxRecords[row] = s.flatDb[s.rowLen*row : s.rowLen*(row+1)]
	// 	}
	// }

	return nil
}

func (s pirServerPunc) answer(q QueryReq, resp *QueryResp) error {
	resp.Answer = make(Row, s.rowLen)
	s.xorRowsFlatSlice(resp.Answer, q.PuncturedSet.Eval())
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
	nRowsRounded := 1 << int(math.Ceil(math.Log2(float64(nRows))/2)*2)
	setSize := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))
	nHints := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))

	return &pirClientPunc{
		nRows:      nRowsRounded,
		setSize:    setSize,
		nHints:     nHints,
		hints:      nil,
		setGen:     NewPRFSetGenerator(source),
		randSource: source,
		servers:    servers,
	}
}

func (c *pirClientPunc) requestHint() (*HintReq, error) {
	c.sets = make([]PuncturableSet, c.nHints)
	for i := 0; i < c.nHints; i++ {
		c.sets[i] = c.setGen.SetGen(c.nRows, c.setSize)
	}
	return &HintReq{
		Sets:          c.sets,
		AuxRecordsSet: c.setGen.SetGen(c.nRows, c.setSize),
	}, nil
}

func (c *pirClientPunc) initHint(resp *HintResp) error {
	c.hints = resp.Hints
	c.auxRecords = resp.AuxRecords
	return nil
}

// Sample a biased coin that comes up heads (true) with
// probability (nHeads/total).
func (c *pirClientPunc) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
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
	i       int
	setIdx  int
	coinBad bool
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
		set = SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
	}

	var puncturedSet, newPuncSet SuccinctSet
	ctx.coinBad = c.bernoulli(c.setSize-1, c.nRows)
	if ctx.coinBad {
		newSet := SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
		newPuncSet = newSet.Punc(c.randomMemberExcept(newSet, i))
		puncturedSet = newPuncSet
	} else {
		puncturedSet = set.Punc(i)
		newSet := SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
		newPuncSet = newSet.Punc(i)
		if ctx.setIdx >= 0 {
			c.sets[ctx.setIdx] = newSet
		}
	}

	return []QueryReq{
			QueryReq{PuncturedSet: newPuncSet},
			QueryReq{PuncturedSet: puncturedSet}},
		ctx
}

func (c *pirClientPunc) auxSetWith(i int) Set {
	puncturedSet := make(Set, 0, len(c.auxRecords))
	for pos, _ := range c.auxRecords {
		puncturedSet = append(puncturedSet, pos)
	}
	if _, present := c.auxRecords[i]; !present {
		replace := c.randSource.Intn(len(puncturedSet))
		delete(c.auxRecords, puncturedSet[replace])
		puncturedSet[replace] = i
	} else {
		delete(c.auxRecords, i)
	}

	return puncturedSet
}

func (c *pirClientPunc) reconstruct(ctx puncQueryCtx, resp []*QueryResp) (Row, error) {
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if ctx.setIdx < 0 {
		return nil, errors.New("couldn't find element in collection")
	} else if ctx.coinBad {
		return nil, errors.New("Bad event coin flip")
	} else if ctx.setIdx == len(c.hints) {
		for _, record := range c.auxRecords {
			if record != nil {
				xorInto(out, record)
			}
		}
		xorInto(out, resp[Right].Answer)
	} else {
		hint := c.hints[ctx.setIdx]
		xorInto(out, hint)
		xorInto(out, resp[Right].Answer)
		// Update hint with refresh info
		xorInto(hint, hint)
		xorInto(hint, resp[Left].Answer)
		xorInto(hint, out)
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
	// TODO: can remove the followong line when bringing back KP's trick
	n = int(math.Min(math.Log(2)*128, float64(len(idxs))))
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

	resps := [][]QueryResp{make([]QueryResp, n), make([]QueryResp, n)}
	err := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for side := range []int{Left, Right} {
		go func(side int) {
			err[side] = c.servers[side].AnswerBatch(reqs[side], &resps[side])
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
	// Read multiple repetitions to handle online errors
	// This can be removed if bringing back KP's trick.
	idxs := make([]int, int(math.Log(2)*128))
	for j := range idxs {
		idxs[j] = i
	}
	vals, err := c.ReadBatch(idxs)
	for j := range idxs {
		if err[j] == nil {
			return vals[j], nil
		}
	}
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
