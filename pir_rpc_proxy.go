package boosted

import (
	"net/rpc"
)

type PirRpcProxy struct {
	remote *rpc.Client

	// Recording requests
	ShouldRecord bool
	HintReqs     []HintReq
	HintResps    []HintResp
	QueryReqs    [][]QueryReq
	QueryResps   [][]QueryResp
}

func NewPirRpcProxy(remote *rpc.Client) *PirRpcProxy {
	registerExtraTypes()
	return &PirRpcProxy{
		remote: remote,
	}
}

func (p *PirRpcProxy) Hint(req *HintReq, resp *HintResp) error {
	err := p.remote.Call("PirRpcServer.Hint", req, &resp)
	if err == nil && p.ShouldRecord {
		p.HintReqs = append(p.HintReqs, *req)
		p.HintResps = append(p.HintResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	err := p.remote.Call("PirRpcServer.AnswerBatch", queries, resps)
	if err == nil && p.ShouldRecord {
		p.QueryReqs = append(p.QueryReqs, queries)
		p.QueryResps = append(p.QueryResps, *resps)
	}
	return err
}
