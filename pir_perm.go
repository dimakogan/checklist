package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type pirPermClient struct {
	nRows   int
	setSize int
	nHints  int

	partition *partition

	hints []Row

	randSource *rand.Rand
}

type pirPermServer struct {
	nRows  int
	rowLen int

	flatDb []byte
}

func NewPirPermClient(src *rand.Rand, nRows int) *pirPermClient {
	setSize := int(math.Sqrt(float64(nRows)))
	nHints := (nRows-1)/setSize + 1
	partition, err := NewPartition(src, nRows, nHints)
	if err != nil {
		panic(fmt.Sprintf("Client failed to create partition: %s", err))
	}
	return &pirPermClient{
		nRows:      nRows,
		setSize:    setSize,
		nHints:     nHints,
		partition:  partition,
		randSource: src,
	}
}

func NewPirPermServer(data []Row) pirPermServer {
	return pirPermServer{
		rowLen: len(data[0]),
		nRows:  len(data),
		flatDb: flattenDb(data),
	}
}

func (s pirPermServer) Hint(req *HintReq, resp *HintResp) error {
	hints := make([]Row, req.NumHints)
	partition, err := NewPartitionFromKey(req.PartitionKey, s.nRows, req.NumHints)
	if err != nil {
		return fmt.Errorf("Server failed to create partition: %s", err)
	}

	for j := range hints {
		hints[j] = make(Row, s.rowLen)
	}

	for i := 0; i < s.nRows; i++ {
		j, _ := partition.Find(i)
		xorInto(hints[j], s.flatDb[s.rowLen*i:s.rowLen*(i+1)])
	}
	resp.Hints = hints

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

func (c *pirPermClient) requestHint() (*HintReq, error) {
	return &HintReq{
		NumHints:     c.nHints,
		PartitionKey: c.partition.Key(),
	}, nil
}

func (c *pirPermClient) initHint(resp *HintResp) error {
	c.hints = resp.Hints
	return nil
}

type permQueryCtx struct {
	i        int
	setIdx   int
	posInSet int
	decoy    int
}

func (c *pirPermClient) query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	setNumber, posInSet := c.partition.Find(i)
	decoy := c.randSource.Intn(c.nRows)
	if decoy != i {
		c.partition.Swap(i, decoy)
	}
	puncSet := c.partition.Set(setNumber)

	return []QueryReq{QueryReq{}, QueryReq{PuncturedSet: &puncSet}},
		func(resp []QueryResp) (Row, error) {
			return c.reconstruct(permQueryCtx{i, setNumber, posInSet, decoy}, resp[1])
		}
}

func (c *pirPermClient) reconstruct(ctx permQueryCtx, resp QueryResp) (Row, error) {
	if ctx.decoy == ctx.i {
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
	newSetNumber, _ := c.partition.Find(ctx.i)
	xorInto(c.hints[newSetNumber], decoyVal)
	xorInto(c.hints[newSetNumber], iVal)

	return iVal, nil
}