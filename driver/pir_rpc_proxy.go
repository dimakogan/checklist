package driver

import (
	"time"

	"checklist/pir"
	"checklist/rpc"
	"checklist/updatable"
)

type RpcProxy struct {
	*rpc.ClientProxy
}

func NewRpcProxy(serverAddr string, useTLS bool, usePersistent bool) (*RpcProxy, error) {
	proxy, err := rpc.NewClientProxy(serverAddr, useTLS, usePersistent, RegisteredTypes())
	if err != nil {
		return nil, err
	}
	return &RpcProxy{proxy}, nil
}

func (p *RpcProxy) KeyUpdates(req updatable.KeyUpdatesReq, resp *updatable.KeyUpdatesResp) error {
	return p.Call("PirServerDriver.KeyUpdates", req, resp)
}

func (p *RpcProxy) Hint(req pir.HintReq, resp *pir.HintResp) error {
	return p.Call("PirServerDriver.Hint", req, resp)
}

func (p *RpcProxy) Answer(query pir.QueryReq, resp *interface{}) error {
	return p.Call("PirServerDriver.Answer", query, resp)
}

func (p *RpcProxy) Configure(config TestConfig, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.Configure", config, &none)
}

func (p *RpcProxy) AddRows(numRows int, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.AddRows", numRows, none)
}

func (p *RpcProxy) DeleteRows(numRows int, none *int) error {
	var non int
	if none == nil {
		none = &non
	}

	return p.Call("PirServerDriver.DeleteRows", numRows, none)
}

func (p *RpcProxy) NumRows(none int, out *int) error {
	return p.Call("PirServerDriver.NumRows", none, out)
}

func (p *RpcProxy) NumKeys(none int, out *int) error {
	return p.Call("PirServerDriver.NumKeys", none, out)
}

func (p *RpcProxy) RowLen(none int, out *int) error {
	return p.Call("PirServerDriver.RowLen", none, out)
}

func (p *RpcProxy) GetRow(idx int, row *RowIndexVal) error {
	return p.Call("PirServerDriver.GetRow", idx, row)
}

func (p *RpcProxy) GetOfflineTimer(none int, out *time.Duration) error {
	return p.Call("PirServerDriver.GetOfflineTimer", none, out)
}

func (p *RpcProxy) GetOnlineTimer(none int, out *time.Duration) error {
	return p.Call("PirServerDriver.GetOnlineTimer", none, out)
}

func (p *RpcProxy) ResetMetrics(none int, none2 *int) error {
	return p.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *RpcProxy) GetOfflineBytes(none int, out *int) error {
	return p.Call("PirServerDriver.GetOfflineBytes", none, out)
}

func (p *RpcProxy) GetOnlineBytes(none int, out *int) error {
	return p.Call("PirServerDriver.GetOnlineBytes", none, out)
}
