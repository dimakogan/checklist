package boosted

import (
	"net/rpc"
	"time"
)

type PirRpcProxy struct {
	remote *rpc.Client

	// Recording requests
	ShouldRecord bool
	HintReqs     []HintReq
	HintResps    []HintResp
	QueryReqs    []QueryReq
	QueryResps   []QueryResp
}

func NewPirRpcProxy(remote *rpc.Client) *PirRpcProxy {
	registerExtraTypes()
	return &PirRpcProxy{
		remote: remote,
	}
}

func (p *PirRpcProxy) Hint(req HintReq, resp *HintResp) error {
	err := p.remote.Call("PirServerDriver.Hint", req, &resp)
	if err == nil && p.ShouldRecord {
		p.HintReqs = append(p.HintReqs, req)
		p.HintResps = append(p.HintResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Answer(query QueryReq, resp *QueryResp) error {
	err := p.remote.Call("PirServerDriver.Answer", query, resp)
	if err == nil && p.ShouldRecord {
		p.QueryReqs = append(p.QueryReqs, query)
		p.QueryResps = append(p.QueryResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Configure(config TestConfig, none *int) error {
	return p.remote.Call("PirServerDriver.Configure", config, none)
}

func (p *PirRpcProxy) AddRows(numRows int, none *int) (err error) {
	return p.remote.Call("PirServerDriver.AddRows", numRows, none)
}

func (p *PirRpcProxy) DeleteRows(numRows int, none *int) (err error) {
	return p.remote.Call("PirServerDriver.DeleteRows", numRows, none)
}

func (p *PirRpcProxy) StartCpuProfile(none int, none2 *int) error {
	return p.remote.Call("PirServerDriver.StartCpuProfile", none, none2)
}

func (p *PirRpcProxy) StopCpuProfile(none int, out *string) error {
	return p.remote.Call("PirServerDriver.StopCpuProfile", none, out)
}

func (p *PirRpcProxy) NumRows(none int, out *int) error {
	return p.remote.Call("PirServerDriver.NumRows", none, out)
}

func (p *PirRpcProxy) GetRow(idx int, row *RowIndexVal) error {
	return p.remote.Call("PirServerDriver.GetRow", idx, row)
}

func (p *PirRpcProxy) GetHintTimer(none int, out *time.Duration) error {
	return p.remote.Call("PirServerDriver.GetHintTimer", none, out)
}

func (p *PirRpcProxy) GetAnswerTimer(none int, out *time.Duration) error {
	return p.remote.Call("PirServerDriver.GetAnswerTimer", none, out)
}

func (p *PirRpcProxy) ResetMetrics(none int, none2 *int) error {
	return p.remote.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *PirRpcProxy) GetHintBytes(none int, out *int) error {
	return p.remote.Call("PirServerDriver.GetHintBytes", none, out)
}

func (p *PirRpcProxy) GetAnswerBytes(none int, out *int) error {
	return p.remote.Call("PirServerDriver.GetAnswerBytes", none, out)
}
