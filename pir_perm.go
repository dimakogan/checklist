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

	hints []Row

	randSource *rand.Rand

	prp     *PRP
	servers [2]PirServer
}

type pirPermServer struct {
	nRows  int
	rowLen int

	flatDb []byte

	randSource *rand.Rand
}

func NewPirPermClient(src *rand.Rand, nRows int, servers [2]PirServer) *pirPermClient {
	setSize := int(math.Sqrt(float64(nRows)))
	return &pirPermClient{
		nRows:      nRows,
		setSize:    setSize,
		nHints:     (nRows-1)/setSize + 1,
		randSource: src,
		servers:    servers,
	}
}

func NewPirPermServer(source *rand.Rand, data []Row) pirPermServer {
	return pirPermServer{
		rowLen:     len(data[0]),
		nRows:      len(data),
		flatDb:     flattenDb(data),
		randSource: source,
	}
}

func numRecordsToUnivSizeBits(nRecords int) int {
	// Round univsize to next power of 4
	return ((int(math.Log2(float64(nRecords)))-1)/2 + 1) * 2
}

func (s pirPermServer) Hint(req *HintReq, resp *HintResp) error {
	hints := make([]Row, req.NumHints)
	prp, err := NewPRP(req.PrpKey, numRecordsToUnivSizeBits(s.nRows))
	if err != nil {
		panic(fmt.Errorf("Failed to create PRP: %s", err))
	}

	for j := range hints {
		hints[j] = make(Row, s.rowLen)
	}

	for i := 0; i < s.nRows; i++ {
		j := prp.Invert(i) / req.NumHints
		xorInto(hints[j], s.flatDb[s.rowLen*i:s.rowLen*(i+1)])
	}
	resp.Hints = hints

	return nil
}

func (s pirPermServer) answer(q QueryReq, resp *QueryResp) error {
	resp.Answer = make(Row, s.rowLen)
	xorRowsFlatSlice(s.flatDb, s.rowLen, q.PuncturedSet, resp.Answer)
	return nil
}

func (s pirPermServer) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	*resps = make([]QueryResp, len(queries))
	for i, q := range queries {
		err := s.answer(q, &(*resps)[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *pirPermClient) Init() error {
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

func (c *pirPermClient) requestHint() (*HintReq, error) {
	univSizeBits := numRecordsToUnivSizeBits(c.nRows)
	key := make([]byte, 16)
	if l, err := c.randSource.Read(key); l != len(key) || err != nil {
		panic(err)
	}
	var err error
	c.prp, err = NewPRP(key, univSizeBits)
	if err != nil {
		return nil, fmt.Errorf("Failed to create PRP: %s", err)
	}

	return &HintReq{
		NumHints: c.nHints,
		PrpKey:   key,
	}, nil
}

func (c *pirPermClient) initHint(resp *HintResp) error {
	c.hints = resp.Hints
	return nil
}

func (c *pirPermClient) Read(i int) (Row, error) {
	queryReq := c.query(i)
	responses := make([]QueryResp, 1)
	err := c.servers[Right].AnswerBatch([]QueryReq{queryReq}, &responses)
	if err != nil {
		return nil, err
	}
	return c.reconstruct(i, responses[0])
}

func (c *pirPermClient) query(i int) QueryReq {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	iPos := c.prp.Invert(i)
	setNumber := iPos / c.setSize
	puncSet := make(Set, c.setSize-1)
	for j := 0; j < c.setSize; j++ {
		pos := c.setSize*setNumber + j
		if pos != iPos {
			puncSet[c.prp.Eval(pos)] = Present_Yes
		}
	}

	return QueryReq{PuncturedSet: puncSet}
}

func (c *pirPermClient) reconstruct(i int, resp QueryResp) (Row, error) {
	out := make(Row, len(c.hints[0]))
	iPos := c.prp.Invert(i)
	setNumber := iPos / c.setSize
	xorInto(out, c.hints[setNumber])
	xorInto(out, resp.Answer)
	return out, nil
}
