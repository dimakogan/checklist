package boosted

import (
	"fmt"
)

type pirClientNonPrivate struct {
	numRows int
	rowLen  int
}

type pirServerNonPrivate struct {
	numRows int
	rowLen  int
	db      []Row
}

func NewPirServerNonPrivate(data []Row) PirServer {
	if len(data) < 1 {
		panic("Database must contain at least one row")
	}

	return &pirServerNonPrivate{
		numRows: len(data),
		rowLen:  len(data[0]),
		db:      data,
	}
}

func (s pirServerNonPrivate) Hint(req HintReq, resp *HintResp) error {
	*resp = HintResp{
		NumRows: s.numRows,
		RowLen:  s.rowLen,
	}
	return nil
}

func (s *pirServerNonPrivate) Answer(q QueryReq, resp *QueryResp) error {
	*resp = QueryResp{Val: s.db[q.Index]}
	return nil
}

func (s *pirServerNonPrivate) NumRows(none int, out *int) error {
	*out = s.numRows
	return nil
}

func (s *pirServerNonPrivate) GetRow(idx int, row *RowIndexVal) error {
	if idx < -1 || idx >= s.numRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.numRows)
	}
	if idx == -1 {
		// return random row
		idx = RandSource().Int() % s.numRows
	}
	row.Value = s.db[idx]
	row.Index = idx
	row.Key = uint32(idx)
	return nil
}

func NewPirClientNonPrivate() *pirClientNonPrivate {
	return &pirClientNonPrivate{}
}

func (c *pirClientNonPrivate) initHint(resp *HintResp) error {
	c.rowLen = resp.RowLen
	c.numRows = resp.NumRows
	return nil
}

func (c *pirClientNonPrivate) query(idx int) ([]QueryReq, ReconstructFunc) {
	queries := make([]QueryReq, 2)
	queries[Left].Index = idx
	queries[Right].Index = idx

	return queries, func(resps []QueryResp) (Row, error) {
		return resps[0].Val, nil
	}
}

func (c *pirClientNonPrivate) dummyQuery() []QueryReq {
	q, _ := c.query(0)
	return q
}
