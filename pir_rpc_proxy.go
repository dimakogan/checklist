package boosted

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func NewPirRpcProxy(serverAddr string) (*PirRpcProxy, error) {
	config := tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", serverAddr, &config)
	if err != nil {
		return nil, fmt.Errorf("client: dial: %s", err)
	}

	io.WriteString(conn, "CONNECT "+rpc.DefaultRPCPath+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected HTTP response: " + resp.Status)
	}
	remote := rpc.NewClient(conn)
	if err != nil {
		return nil, err
	}
	registerExtraTypes()
	return &PirRpcProxy{
		remote: remote,
	}, nil
}

func (p *PirRpcProxy) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	return p.remote.Call("PirServerDriver.KeyUpdates", req, resp)
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

func (p *PirRpcProxy) GetOfflineTimer(none int, out *time.Duration) error {
	return p.remote.Call("PirServerDriver.GetOfflineTimer", none, out)
}

func (p *PirRpcProxy) GetOnlineTimer(none int, out *time.Duration) error {
	return p.remote.Call("PirServerDriver.GetOnlineTimer", none, out)
}

func (p *PirRpcProxy) ResetMetrics(none int, none2 *int) error {
	return p.remote.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *PirRpcProxy) GetOfflineBytes(none int, out *int) error {
	return p.remote.Call("PirServerDriver.GetOfflineBytes", none, out)
}

func (p *PirRpcProxy) GetOnlineBytes(none int, out *int) error {
	return p.remote.Call("PirServerDriver.GetOnlineBytes", none, out)
}
