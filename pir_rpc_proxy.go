package boosted

import (
	"time"

	"github.com/dimakogan/boosted-pir/rpc"
)

type PirRpcProxy struct {
	*rpc.ClientProxy
}

func NewPirRpcProxy(serverAddr string, useTLS bool, usePersistent bool) (*PirRpcProxy, error) {
	proxy, err := rpc.NewClientProxy(serverAddr, useTLS, usePersistent)
	if err != nil {
		return nil, err
	}
	return &PirRpcProxy{proxy}, nil
}

func (p *PirRpcProxy) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	return p.Call("PirServerDriver.KeyUpdates", req, resp)
}

func (p *PirRpcProxy) Hint(req HintReq, resp *HintResp) error {
	return p.Call("PirServerDriver.Hint", req, &resp)
}

func (p *PirRpcProxy) Answer(query QueryReq, resp *QueryResp) error {
	return p.Call("PirServerDriver.Answer", query, resp)
}

func (p *PirRpcProxy) Configure(config TestConfig, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.Configure", config, &none)
}

func (p *PirRpcProxy) AddRows(numRows int, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.AddRows", numRows, none)
}

func (p *PirRpcProxy) DeleteRows(numRows int, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.DeleteRows", numRows, none)
}

func (p *PirRpcProxy) NumRows(none int, out *int) error {
	return p.Call("PirServerDriver.NumRows", none, out)
}

func (p *PirRpcProxy) GetRow(idx int, row *RowIndexVal) error {
	return p.Call("PirServerDriver.GetRow", idx, row)
}

func (p *PirRpcProxy) GetOfflineTimer(none int, out *time.Duration) error {
	return p.Call("PirServerDriver.GetOfflineTimer", none, out)
}

func (p *PirRpcProxy) GetOnlineTimer(none int, out *time.Duration) error {
	return p.Call("PirServerDriver.GetOnlineTimer", none, out)
}

func (p *PirRpcProxy) ResetMetrics(none int, none2 *int) error {
	return p.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *PirRpcProxy) GetOfflineBytes(none int, out *int) error {
	return p.Call("PirServerDriver.GetOfflineBytes", none, out)
}

func (p *PirRpcProxy) GetOnlineBytes(none int, out *int) error {
	return p.Call("PirServerDriver.GetOnlineBytes", none, out)
}
