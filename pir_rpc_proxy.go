package boosted

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ugorji/go/codec"
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

type httpPostCodec struct {
	http        *http.Client
	requestPath string
	encoder     *codec.Encoder
	decoder     *codec.Decoder
	bodyReader  chan (io.ReadCloser)
	bodyCloser  io.Closer
}

func newHttpPostCodec(serverAddr string) *httpPostCodec {
	config := tls.Config{
		InsecureSkipVerify: true,
	}

	http := &http.Client{
		Transport: &http.Transport{
			DialTLS: func(network, addr string) (net.Conn, error) {
				return tls.Dial("tcp", addr, &config)
			},
		},
	}

	return &httpPostCodec{
		http:        http,
		requestPath: "https://" + serverAddr + rpc.DefaultRPCPath,
		encoder:     codec.NewEncoderBytes(nil, CodecHandle()),
		decoder:     codec.NewDecoder(nil, CodecHandle()),
		bodyReader:  make(chan io.ReadCloser),
	}
}

func (c *httpPostCodec) WriteRequest(req *rpc.Request, body interface{}) error {
	var reqBuf []byte
	c.encoder.ResetBytes(&reqBuf)
	err := c.encoder.Encode(req)
	if err != nil {
		return fmt.Errorf("encoder WriteRequest header failed: %v", err)
	}
	err = c.encoder.Encode(body)
	if err != nil {
		return fmt.Errorf("encoder WriteRequest body failed: %v", err)
	}
	resp, err := c.http.Post(c.requestPath, "application/octet-stream", bytes.NewBuffer(reqBuf))
	if err != nil {
		return fmt.Errorf("failed HTTP Get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed HTTP Get: %v", resp.StatusCode)
	}
	c.bodyReader <- resp.Body
	return err
}

func (c *httpPostCodec) ReadResponseHeader(header *rpc.Response) error {
	respBody := <-c.bodyReader
	c.decoder.Reset(respBody)
	c.bodyCloser = respBody
	return c.decoder.Decode(header)
}

func (c *httpPostCodec) ReadResponseBody(body interface{}) error {
	defer c.bodyCloser.Close()
	return c.decoder.Decode(body)
}

func (c *httpPostCodec) Close() error {
	c.http.CloseIdleConnections()
	return nil
}

func NewHTTPSRPCClient(serverAddr string) (*rpc.Client, error) {
	return rpc.NewClientWithCodec(newHttpPostCodec(serverAddr)), nil
}

func NewTCPRPCClient(serverAddr string) (*rpc.Client, error) {
	var err error
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(codec.GoRpc.ClientCodec(conn, CodecHandle())), nil
}

func NewPirRpcProxy(serverAddr string, useTLS bool, usePersistent bool) (*PirRpcProxy, error) {
	proxy := PirRpcProxy{serverAddr: serverAddr, useTLS: useTLS}
	var err error
	if usePersistent {
		if proxy.remote, err = proxy.connect(); err != nil {
			return nil, err
		}
	}
	return &proxy, nil
}

func (p *PirRpcProxy) connect() (*rpc.Client, error) {
	if p.useTLS {
		return NewHTTPSRPCClient(p.serverAddr)
	} else {
		return NewTCPRPCClient(p.serverAddr)
	}
}

func (p *PirRpcProxy) Close() {
	if p.remote != nil {
		p.remote.Close()
	}
}

func (p *PirRpcProxy) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}

	return remote.Call("PirServerDriver.KeyUpdates", req, resp)
}

func (p *PirRpcProxy) Hint(req HintReq, resp *HintResp) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	err := remote.Call("PirServerDriver.Hint", req, &resp)
	if err == nil && p.ShouldRecord {
		p.HintReqs = append(p.HintReqs, req)
		p.HintResps = append(p.HintResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Answer(query QueryReq, resp *QueryResp) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}

	err := remote.Call("PirServerDriver.Answer", query, resp)
	if err == nil && p.ShouldRecord {
		p.QueryReqs = append(p.QueryReqs, query)
		p.QueryResps = append(p.QueryResps, *resp)
	}
	return err
}

func (p *PirRpcProxy) Configure(config TestConfig, none *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	var non int
	if none == nil {
		none = &non
	}
	return remote.Call("PirServerDriver.Configure", config, &none)
}

func (p *PirRpcProxy) AddRows(numRows int, none *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	var non int
	if none == nil {
		none = &non
	}
	return remote.Call("PirServerDriver.AddRows", numRows, none)
}

func (p *PirRpcProxy) DeleteRows(numRows int, none *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	var non int
	if none == nil {
		none = &non
	}
	return remote.Call("PirServerDriver.DeleteRows", numRows, none)
}

func (p *PirRpcProxy) StartCpuProfile(none int, none2 *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.StartCpuProfile", none, none2)
}

func (p *PirRpcProxy) StopCpuProfile(none int, out *string) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.StopCpuProfile", none, out)
}

func (p *PirRpcProxy) NumRows(none int, out *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.NumRows", none, out)
}

func (p *PirRpcProxy) GetRow(idx int, row *RowIndexVal) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.GetRow", idx, row)
}

func (p *PirRpcProxy) GetOfflineTimer(none int, out *time.Duration) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.GetOfflineTimer", none, out)
}

func (p *PirRpcProxy) GetOnlineTimer(none int, out *time.Duration) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.GetOnlineTimer", none, out)
}

func (p *PirRpcProxy) ResetMetrics(none int, none2 *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.ResetMetrics", none, none2)
}

func (p *PirRpcProxy) GetOfflineBytes(none int, out *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.GetOfflineBytes", none, out)
}

func (p *PirRpcProxy) GetOnlineBytes(none int, out *int) error {
	remote := p.remote
	if remote == nil {
		var err error
		if remote, err = p.connect(); err != nil {
			return err
		}
		defer remote.Close()
	}
	return remote.Call("PirServerDriver.GetOnlineBytes", none, out)
}
