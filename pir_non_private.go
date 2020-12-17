package boosted

import (
	"fmt"
)

type pirClientNonPrivate struct {
	nRows  int
	rowLen int
}

type pirServerNonPrivate struct {
	nRows  int
	rowLen int
	db     []Row
}

func NewPirServerNonPrivate(data []Row) PirDB {
	if len(data) < 1 {
		panic("Database must contain at least one row")
	}

	return &pirServerNonPrivate{
		nRows:  len(data),
		rowLen: len(data[0]),
		db:     data,
	}
}

func (s pirServerNonPrivate) Hint(req HintReq, resp *HintResp) error {
	*resp = HintResp{
		NumRows: s.nRows,
		RowLen:  s.rowLen,
	}
	return nil
}

func (s *pirServerNonPrivate) Answer(q QueryReq, resp *QueryResp) error {
	*resp = QueryResp{Val: s.db[q.Index]}
	return nil
}

func (s *pirServerNonPrivate) NumRows(none int, out *int) error {
	*out = s.nRows
	return nil
}

func (s *pirServerNonPrivate) GetRow(idx int, row *RowIndexVal) error {
	if idx < -1 || idx >= s.nRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, s.nRows)
	}
	if idx == -1 {
		// return random row
		idx = RandSource().Int() % s.nRows
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
	c.nRows = resp.NumRows
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
