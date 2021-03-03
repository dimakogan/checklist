package boosted

type pirClientNonPrivate struct {
	nRows  int
	rowLen int
}

type pirServerNonPrivate struct {
	*staticDB
}

func NewPirServerNonPrivate(db *staticDB) *pirServerNonPrivate {
	if db.numRows < 1 {
		panic("Database must contain at least one row")
	}

	return &pirServerNonPrivate{db}
}

func (s pirServerNonPrivate) Hint(req HintReq, resp *HintResp) error {
	*resp = HintResp{
		NumRows: s.numRows,
		RowLen:  s.rowLen,
	}
	return nil
}

func (s *pirServerNonPrivate) Answer(q QueryReq, resp *QueryResp) error {
	*resp = QueryResp{Val: s.flatDb[q.Index*s.rowLen : (q.Index+1)*s.rowLen]}
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
