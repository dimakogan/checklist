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

	servers [2]PirServer
}

type pirPermServer struct {
	nRows  int
	rowLen int

	flatDb []byte
}

func NewPirPermClient(src *rand.Rand, nRows int, servers [2]PirServer) *pirPermClient {
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
		servers:    servers,
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
		j := partition.Find(i)
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
	return &HintReq{
		NumHints:     c.nHints,
		PartitionKey: c.partition.Key(),
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
	setNumber := c.partition.Find(i)
	puncSet := c.partition.Set(setNumber)
	delete(puncSet, i)

	return QueryReq{PuncturedSet: puncSet}
}

func (c *pirPermClient) reconstruct(i int, resp QueryResp) (Row, error) {
	out := make(Row, len(c.hints[0]))
	setNumber := c.partition.Find(i)
	xorInto(out, c.hints[setNumber])
	xorInto(out, resp.Answer)
	return out, nil
}
