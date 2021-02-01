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
	serverAddr string
	useTLS     bool
	remote     *rpc.Client

	// Recording requests
	ShouldRecord bool
	HintReqs     []HintReq
	HintResps    []HintResp
	QueryReqs    []QueryReq
	QueryResps   []QueryResp
}

func NewHTTPSRPCClient(serverAddr string) (*rpc.Client, error) {
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
	return remote, nil
}

func NewPirRpcProxy(serverAddr string, useTLS bool, usePersistent bool) (*PirRpcProxy, error) {
	proxy := PirRpcProxy{serverAddr: serverAddr, useTLS: true}
	var err error
	if usePersistent {
		if proxy.remote, err = proxy.getRemote(); err != nil {
			return nil, err
		}
	}
	return &proxy, nil
}

func (p *PirRpcProxy) getRemote() (*rpc.Client, error) {
	if p.remote != nil {
		return p.remote, nil
	}
	if p.useTLS {
		return NewHTTPSRPCClient(p.serverAddr)
	} else {
		return rpc.DialHTTP("tcp", p.serverAddr)
	}
}

func (p *PirRpcProxy) Close() {
	if p.remote != nil {
		p.remote.Close()
	}
}

func (p *PirRpcProxy) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.KeyUpdates", req, resp)
}

func (p *PirRpcProxy) Hint(req HintReq, resp *HintResp) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	err = remote.Call("PirServerDriver.Hint", req, &resp)
	if err == nil && p.ShouldRecord {
		p.HintReqs = append(p.HintReqs, req)
		p.HintResps = append(p.HintResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Answer(query QueryReq, resp *QueryResp) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	err = remote.Call("PirServerDriver.Answer", query, resp)
	if err == nil && p.ShouldRecord {
		p.QueryReqs = append(p.QueryReqs, query)
		p.QueryResps = append(p.QueryResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Configure(config TestConfig, none *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.Configure", config, none)
}

func (p *PirRpcProxy) AddRows(numRows int, none *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.AddRows", numRows, none)
}

func (p *PirRpcProxy) DeleteRows(numRows int, none *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.DeleteRows", numRows, none)
}

func (p *PirRpcProxy) StartCpuProfile(none int, none2 *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.StartCpuProfile", none, none2)
}

func (p *PirRpcProxy) StopCpuProfile(none int, out *string) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.StopCpuProfile", none, out)
}

func (p *PirRpcProxy) NumRows(none int, out *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.NumRows", none, out)
}

func (p *PirRpcProxy) GetRow(idx int, row *RowIndexVal) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.GetRow", idx, row)
}

func (p *PirRpcProxy) GetOfflineTimer(none int, out *time.Duration) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.GetOfflineTimer", none, out)
}

func (p *PirRpcProxy) GetOnlineTimer(none int, out *time.Duration) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.GetOnlineTimer", none, out)
}

func (p *PirRpcProxy) ResetMetrics(none int, none2 *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *PirRpcProxy) GetOfflineBytes(none int, out *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.GetOfflineBytes", none, out)
}

func (p *PirRpcProxy) GetOnlineBytes(none int, out *int) error {
	var remote *rpc.Client
	var err error
	if remote, err = p.getRemote(); err != nil {
		return err
	}
	defer remote.Close()
	return remote.Call("PirServerDriver.GetOnlineBytes", none, out)
}
