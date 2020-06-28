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

	keys       []*SetKey
	hints      []Row
	auxRecords map[int]Row

	randSource *rand.Rand

	server PuncPirServer
}

type pirServerPunc struct {
	nRows  int
	rowLen int

	flatDb []byte

	randSource *rand.Rand
}

type PuncPirServer interface {
	Hint(req *HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *QueryResp) error
	AnswerBatch(q []QueryReq, resp []QueryResp) error
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

func (s *pirServerPunc) xorRowsFlatSlice(out Row, rows Set) int {
	bytes := 0
	//	setS := setToSlice(set)

	for row, _ := range rows {
		if row >= s.nRows {
			continue
		}
		xorInto(out, s.flatDb[s.rowLen*row:s.rowLen*(row+1)])
		bytes += s.rowLen
	}
	return bytes
}

func NewPirServerPunc(source *rand.Rand, data []Row) *pirServerPunc {
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
		nRows:      len(data),
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
	nHints := len(req.Sets)
	hints := make([]Row, nHints)

	bytes := 0
	for j := 0; j < nHints; j++ {
		hints[j] = make(Row, s.rowLen)
		set := req.Sets[j].Eval()
		bytes = bytes + s.xorRowsFlatSlice(hints[j], set)
	}
	resp.Hints = hints

	auxSet := req.AuxRecordsSet.Eval()
	resp.AuxRecords = make(map[int]Row)
	for row := range auxSet {
		if row < s.nRows {
			resp.AuxRecords[row] = s.flatDb[s.rowLen*row : s.rowLen*(row+1)]
		}
	}

	return nil
}

func (s *pirServerPunc) Answer(q QueryReq, resp *QueryResp) error {
	resp.Answer = make(Row, s.rowLen)
	s.xorRowsFlatSlice(resp.Answer, q.PuncturedSet)
	return nil
}

func (s *pirServerPunc) AnswerBatch(queries []QueryReq, resps []QueryResp) error {
	for i, q := range queries {
		err := s.Answer(q, &resps[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func NewPirClientPunc(source *rand.Rand, nRows int, server PuncPirServer) *pirClientPunc {
	// TODO: Maybe better to just do this with integer ops.
	nRowsRounded := 1 << int(math.Ceil(math.Log2(float64(nRows))/2)*2)
	setSize := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))
	nHints := int(math.Round(math.Pow(float64(nRowsRounded), 0.5)))

	return &pirClientPunc{
		nRows:      nRowsRounded,
		setSize:    setSize,
		nHints:     nHints,
		hints:      nil,
		randSource: source,
		server:     server,
	}
}

func (c *pirClientPunc) requestHint() (*HintReq, error) {
	c.keys = make([]*SetKey, c.nHints)
	for i := 0; i < c.nHints; i++ {
		c.keys[i] = SetGen(c.randSource, c.nRows, c.setSize)
	}
	return &HintReq{
		Sets:          c.keys,
		AuxRecordsSet: SetGen(c.randSource, c.nRows, c.setSize),
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
	for j, key := range c.keys {
		if key.InSet(i) {
			return j
		}
	}
	return -1
}

func (c *pirClientPunc) query(i int) ([]QueryReq, int) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}

	var key *SetKey
	querySetIdx := 0
	if querySetIdx = c.findIndex(i); querySetIdx >= 0 {
		key = c.keys[querySetIdx]
	} else {
		key = SetGenWith(RandSource(), c.nRows, c.setSize, i)
	}

	var puncSet, newPuncSet Set
	coin := c.bernoulli(c.setSize-1, c.nRows)
	if coin {
		newSet := SetGenWith(RandSource(), c.nRows, c.setSize, i)
		querySetIdx = -1
		newPuncSet = newSet.Punc(newSet.RandomMemberExcept(c.randSource, i))
		puncSet = newPuncSet

	} else {
		puncSet = key.Punc(i)
		newSet := SetGenWith(RandSource(), c.nRows, c.setSize, i)
		newPuncSet = newSet.Punc(i)
		if querySetIdx >= 0 {
			c.keys[querySetIdx] = newSet
		}
	}

	return []QueryReq{
		QueryReq{PuncturedSet: puncSet},
		QueryReq{PuncturedSet: newPuncSet},
	}, querySetIdx
}

func (c *pirClientPunc) auxSetWith(i int) Set {
	puncSet := make(Set)
	for pos, _ := range c.auxRecords {
		puncSet[pos] = Present_Yes
	}
	if _, present := puncSet[i]; !present {
		var remove int
		remove = puncSet.RandomMember(c.randSource)
		delete(puncSet, remove)
		delete(c.auxRecords, remove)
		puncSet[i] = Present_Yes
	} else {
		delete(c.auxRecords, i)
	}
	return puncSet
}

func (c *pirClientPunc) reconstruct(querySetIdx int, resp []*QueryResp) (Row, error) {
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if querySetIdx < 0 {
		return nil, errors.New("couldn't find element in collection")
	} else if querySetIdx == len(c.hints) {
		for _, record := range c.auxRecords {
			if record != nil {
				xorInto(out, record)
			}
		}
		xorInto(out, resp[0].Answer)
	} else {
		xorInto(out, c.hints[querySetIdx])
		xorInto(out, resp[0].Answer)
		// Update hint with refresh info
		xorInto(c.hints[querySetIdx], c.hints[querySetIdx])
		xorInto(c.hints[querySetIdx], resp[1].Answer)
		xorInto(c.hints[querySetIdx], out)
	}

	return out, nil
}

func (c *pirClientPunc) Init() error {
	hintReq, err := c.requestHint()
	if err != nil {
		return err
	}
	var hintResp HintResp
	err = c.server.Hint(hintReq, &hintResp)
	if err != nil {
		return err
	}
	return c.initHint(&hintResp)
}

func (c *pirClientPunc) ReadBatch(idxs []int) ([]Row, []error) {
	leftReqs := make([]QueryReq, len(idxs))
	rightReqs := make([]QueryReq, len(idxs))
	errs := make([]error, len(idxs))
	querySetIdxs := make([]int, len(idxs))

	for pos, i := range idxs {
		var queryReqs []QueryReq
		queryReqs, querySetIdxs[pos] = c.query(i)
		leftReqs[pos] = queryReqs[0]
		rightReqs[pos] = queryReqs[1]
	}

	leftResps := make([]QueryResp, len(idxs))
	rightResps := make([]QueryResp, len(idxs))
	var rightErr, leftErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		leftErr = c.server.AnswerBatch(leftReqs, leftResps)
		wg.Done()
	}()
	go func() {
		rightErr = c.server.AnswerBatch(rightReqs, rightResps)
		wg.Done()
	}()
	wg.Wait()

	vals := make([]Row, len(idxs))
	for i := 0; i < len(idxs); i++ {
		if leftErr != nil {
			errs[i] = leftErr
			continue
		} else if rightErr != nil {
			errs[i] = rightErr
			continue
		}

		vals[i], errs[i] = c.reconstruct(querySetIdxs[i], []*QueryResp{&leftResps[i], &rightResps[i]})
	}
	return vals, errs
}

func (c *pirClientPunc) Read(i int) (Row, error) {
	queryReq, querySetIdx := c.query(i)

	var queryResp QueryResp
	err := c.server.Answer(queryReq[0], &queryResp)
	if err != nil {
		return nil, fmt.Errorf("Error on query Answer: %w", err)
	}
	var refreshResp QueryResp
	err = c.server.Answer(queryReq[1], &refreshResp)
	if err != nil {
		return nil, fmt.Errorf("Error on refresh Answer: %w", err)
	}
	val, err := c.reconstruct(querySetIdx, []*QueryResp{&queryResp, &refreshResp})
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *pirClientPunc) NumCovered() int {
	covered := make(map[int]bool)
	for _, key := range c.keys {
		set := key.Eval()
		for elem, _ := range set {
			covered[elem] = true
		}
	}
	return len(covered)
}
