package boosted

import (
	"net/rpc"
)

type PirRpcProxy struct {
	remote *rpc.Client
}

func NewPirRpcProxy(remote *rpc.Client) *PirRpcProxy {
	registerExtraTypes()
	return &PirRpcProxy{
		remote: remote,
	}
}

func (p PirRpcProxy) Hint(req *HintReq, resp *HintResp) error {
	return p.remote.Call("PirRpcServer.Hint", req, &resp)
}

func (p PirRpcProxy) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	return p.remote.Call("PirRpcServer.AnswerBatch", queries, resps)
}
