package pir

import (
	"math/rand"
)

type nonPrivateClient struct {
}

type NonPrivateHintReq struct {
}

type NonPrivateHintResp struct {
	NRows int
}

func NewNonPrivateHintReq() *NonPrivateHintReq {
	var req NonPrivateHintReq
	return &req
}

func (req *NonPrivateHintReq) Process(db StaticDB) (HintResp, error) {
	return &NonPrivateHintResp{db.NumRows}, nil
}

func (resp *NonPrivateHintResp) InitClient(source *rand.Rand) Client {
	return &nonPrivateClient{}
}

func (resp *NonPrivateHintResp) NumRows() int {
	return resp.NRows
}

type NonPrivateQueryReq struct {
	Index int
}
type NonPrivateQueryResp struct {
	Row
}

func (req *NonPrivateQueryReq) Process(db StaticDB) (interface{}, error) {
	idx := req.Index
	return &NonPrivateQueryResp{db.FlatDb[idx*db.RowLen : (idx+1)*db.RowLen]}, nil
}

func NewClientNonPrivate() *nonPrivateClient {
	return &nonPrivateClient{}
}

func (c *nonPrivateClient) Query(idx int) ([]QueryReq, ReconstructFunc) {
	var q NonPrivateQueryReq
	q = NonPrivateQueryReq{idx}
	return []QueryReq{&q, &q}, func(resps []interface{}) (Row, error) {
		return (resps[0].(*NonPrivateQueryResp)).Row, nil
	}
}

func (c *nonPrivateClient) DummyQuery() []QueryReq {
	q, _ := c.Query(0)
	return q
}

func (c *nonPrivateClient) StateSize() (int, int) {
	return 0, 0
}
