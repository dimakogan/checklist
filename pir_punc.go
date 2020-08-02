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

	keys       []SetKey
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
		setGen:     NewPrpSetGenerator(source),
		randSource: source,
		servers:    servers,
	}
}

func (c *pirClientPunc) requestHint() (*HintReq, error) {
	c.keys = make([]SetKey, c.nHints)
	for i := 0; i < c.nHints; i++ {
		c.keys[i] = c.setGen.SetGen(c.nRows, c.setSize)
	}
	return &HintReq{
		Sets:          c.keys,
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

	var key SetKey
	querySetIdx := 0
	if querySetIdx = c.findIndex(i); querySetIdx >= 0 {
		key = c.keys[querySetIdx]
	} else {
		key = SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
	}

	var puncSet, newPuncSet SetKey
	coin := c.bernoulli(c.setSize-1, c.nRows)
	if coin {
		newSet := SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
		querySetIdx = -1
		newPuncSet = newSet.Punc(c.randomMemberExcept(newSet, i))
		puncSet = newPuncSet

	} else {
		puncSet = key.Punc(i)
		newSet := SetGenWith(c.setGen, c.randSource, c.nRows, c.setSize, i)
		newPuncSet = newSet.Punc(i)
		if querySetIdx >= 0 {
			c.keys[querySetIdx] = newSet
		}
	}

	return []QueryReq{
			QueryReq{PuncturedSet: newPuncSet},
			QueryReq{PuncturedSet: puncSet}},
		querySetIdx
}

func (c *pirClientPunc) auxSetWith(i int) Set {
	puncSet := make(Set, 0, len(c.auxRecords))
	for pos, _ := range c.auxRecords {
		puncSet = append(puncSet, pos)
	}
	if _, present := c.auxRecords[i]; !present {
		replace := c.randSource.Intn(len(puncSet))
		delete(c.auxRecords, puncSet[replace])
		puncSet[replace] = i
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
		xorInto(out, resp[Right].Answer)
	} else {
		xorInto(out, c.hints[querySetIdx])
		xorInto(out, resp[Right].Answer)
		// Update hint with refresh info
		xorInto(c.hints[querySetIdx], c.hints[querySetIdx])
		xorInto(c.hints[querySetIdx], resp[Left].Answer)
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
	err = c.servers[Left].Hint(hintReq, &hintResp)
	if err != nil {
		return err
	}
	return c.initHint(&hintResp)
}

func (c *pirClientPunc) ReadBatch(idxs []int) ([]Row, []error) {
	return c.ReadBatchAtLeast(idxs, len(idxs))
}

// TODO: need to bring back KP's trick since as is, this does not send the query
// in the bad coin-flip case, thus leaking.
func (c *pirClientPunc) ReadBatchAtLeast(idxs []int, n int) ([]Row, []error) {
	reqs := [][]QueryReq{make([]QueryReq, n), make([]QueryReq, n)}
	querySetIdxs := make([]int, len(idxs))

	nOk := 0
	for pos, i := range idxs {
		var queryReqs []QueryReq
		queryReqs, querySetIdxs[pos] = c.query(i)
		if querySetIdxs[pos] >= 0 {
			reqs[Left][nOk] = queryReqs[Left]
			reqs[Right][nOk] = queryReqs[Right]
			nOk++
		}
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
		if querySetIdxs[i] >= 0 && nOk < n {
			vals[i], errs[i] = c.reconstruct(querySetIdxs[i], []*QueryResp{&resps[Left][nOk], &resps[Right][nOk]})
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
	for _, key := range c.keys {
		set := key.Eval()
		for _, elem := range set {
			covered[elem] = true
		}
	}
	return len(covered)
}

// Sample a random element of the set that is not equal to `idx`.
func (c *pirClientPunc) randomMemberExcept(key SetKey, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := key.ElemAt(c.randSource.Intn(c.setSize))
		if val != idx {
			return val
		}
	}
}

func (c *pirClientPunc) setGenWith(univSize int, setSize int, val int) SetKey {
	return nil
}
