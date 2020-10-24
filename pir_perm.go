package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type pirPermClient struct {
	nRows int

	partition *partition

	hints []Row

	randSource *rand.Rand
}

type pirPermServer struct {
	nRows  int
	rowLen int

	flatDb []byte
}

func NewPirPermClient(src *rand.Rand) *pirPermClient {
	return &pirPermClient{randSource: src}
}

func NewPirPermServer(data []Row) pirPermServer {
	return pirPermServer{
		rowLen: len(data[0]),
		nRows:  len(data),
		flatDb: flattenDb(data),
	}
}

func (s pirPermServer) Hint(req HintReq, resp *HintResp) error {
	setSize := int(math.Sqrt(float64(s.nRows)))
	nHints := (s.nRows-1)/setSize + 1
	src := rand.New(rand.NewSource(req.RandSeed))
	key := make([]byte, 16)
	if l, err := src.Read(key); l != len(key) || err != nil {
		panic(err)
	}

	partition, err := NewPartition(key, s.nRows, nHints)
	if err != nil {
		panic(fmt.Sprintf("Client failed to create partition: %s", err))
	}

	hints := make([]Row, nHints)

	for j := range hints {
		hints[j] = make(Row, s.rowLen)
	}

	for i := 0; i < s.nRows; i++ {
		j, _ := partition.Find(i)
		xorInto(hints[j], s.flatDb[s.rowLen*i:s.rowLen*(i+1)])
	}

	resp.NumRows = s.nRows
	resp.Hints = hints
	resp.SetGenKey = partition.Key()

	return nil
}

func (s pirPermServer) Answer(q QueryReq, resp *QueryResp) error {
	if q.PuncturedSet == nil {
		return nil
	}
	resp.Values = make([]Row, 0, q.PuncturedSet.Size())
	for _, row := range q.PuncturedSet.Eval() {
		if row < s.nRows {
			resp.Values = append(resp.Values, s.flatDb[s.rowLen*row:s.rowLen*(row+1)])
		} else {
			resp.Values = append(resp.Values, nil)
		}
	}
	return nil
}

func (s pirPermServer) GetRow(idx int, row *RowIndexVal) error {
	if idx < 0 || idx >= s.nRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.nRows)
	}
	row.Value = s.flatDb[idx*s.rowLen : (idx+1)*s.rowLen]
	row.Index = idx
	row.Key = uint32(idx)
	return nil
}

func (c *pirPermClient) initHint(resp *HintResp) (err error) {
	c.hints = resp.Hints
	c.nRows = resp.NumRows
	c.partition, err = NewPartition(resp.SetGenKey, resp.NumRows, len(c.hints))
	if err != nil {
		return fmt.Errorf("Server failed to create partition: %s", err)
	}

	return nil
}

type permQueryCtx struct {
	i         int
	setIdx    int
	posInSet  int
	decoy     int
	newSetIdx int
}

func (c *pirPermClient) query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	setNumber, posInSet := c.partition.Find(i)
	decoy := c.randSource.Intn(c.nRows)
	newSetIdx, _ := c.partition.Find(i)
	if setNumber != newSetIdx {
		c.partition.Swap(i, decoy)
	}
	puncSet := c.partition.Set(setNumber)

	return []QueryReq{{}, {PuncturedSet: &puncSet}},
		func(resp []QueryResp) (Row, error) {
			return c.reconstruct(permQueryCtx{i, setNumber, posInSet, decoy, newSetIdx}, resp[1])
		}
}

func (c *pirPermClient) dummyQuery() []QueryReq {
	set := c.partition.Set(0)
	puncSet := set.Punc(set.ElemAt(0))
	return []QueryReq{{}, {PuncturedSet: puncSet}}
}

func (c *pirPermClient) reconstruct(ctx permQueryCtx, resp QueryResp) (Row, error) {
	if ctx.setIdx == ctx.newSetIdx {
		return resp.Values[ctx.posInSet], nil
	}
	decoyVal := resp.Values[ctx.posInSet]
	iVal := make(Row, len(c.hints[ctx.setIdx]))
	xorInto(iVal, c.hints[ctx.setIdx])
	for pos, val := range resp.Values {
		if pos != ctx.posInSet {
			xorInto(iVal, val)
		}
	}

	xorInto(c.hints[ctx.newSetIdx], decoyVal)
	xorInto(c.hints[ctx.newSetIdx], iVal)

	return iVal, nil
}
