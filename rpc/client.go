package rpc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"reflect"

	"github.com/ugorji/go/codec"
)

type ClientProxy struct {
	serverAddr string
	useTLS     bool
	persistent bool

	codecHandle codec.Handle

	// Cached
	cachedCodec  rpc.ClientCodec
	cachedClient *rpc.Client

	// Recording requests
	shouldRecord bool
	RecordedReqs []RecordedRequest
}

type httpPostCodec struct {
	http       *http.Client
	serverAddr string
	encoder    *codec.Encoder
	decoder    *codec.Decoder
	bodyReader chan (io.ReadCloser)
	bodyCloser io.Closer
}

func newHttpPostCodec(codecHandle codec.Handle, serverAddr string, usePersistent bool) *httpPostCodec {
	config := tls.Config{
		InsecureSkipVerify: true,
	}

	http := &http.Client{
		Transport: &http.Transport{
			DialTLS: func(network, addr string) (net.Conn, error) {
				return tls.Dial("tcp", addr, &config)
			},
			DisableKeepAlives: !usePersistent,
		},
	}

	return &httpPostCodec{
		http:       http,
		serverAddr: serverAddr,
		encoder:    codec.NewEncoderBytes(nil, codecHandle),
		decoder:    codec.NewDecoder(nil, codecHandle),
		bodyReader: make(chan io.ReadCloser),
	}
}

func (c *httpPostCodec) WriteRequest(rpcReq *rpc.Request, body interface{}) error {
	var reqBuf []byte
	c.encoder.ResetBytes(&reqBuf)
	err := c.encoder.Encode(rpcReq)
	if err != nil {
		return fmt.Errorf("encoder WriteRequest header failed: %v", err)
	}
	err = c.encoder.Encode(body)
	if err != nil {
		return fmt.Errorf("encoder WriteRequest body failed: %v", err)
	}

	path := rpc.DefaultRPCPath + "/" + rpcReq.ServiceMethod
	url := "https://" + c.serverAddr + path
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBuf))
	httpReq.Header.Set("Content-Type", "application/octet-stream")
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed HTTP Get: %v", err)
	}
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("failed HTTP POST: %v", httpResp.StatusCode)
	}
	c.bodyReader <- httpResp.Body
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

type RecordedRequest struct {
	Method   string
	ReqBody  interface{}
	Error    error
	RespBody interface{}
}

type recorderCodec struct {
	nextCodec    rpc.ClientCodec
	recordedReqs []RecordedRequest

	lastSeq uint64
}

func NewClientProxy(serverAddr string, useTLS bool, usePersistent bool, types []reflect.Type) (*ClientProxy, error) {
	proxy := ClientProxy{serverAddr: serverAddr, useTLS: useTLS, codecHandle: CodecHandle(types)}
	if usePersistent || useTLS {
		// Always cache TLS codec
		codec, err := proxy.codec()
		if err != nil {
			return nil, err
		}
		proxy.cachedCodec = codec

		proxy.cachedClient, err = proxy.rpcClient()
		if err != nil {
			proxy.cachedCodec.Close()
			return nil, err
		}
		proxy.persistent = true
	}
	return &proxy, nil
}

func (p *ClientProxy) codec() (rpc.ClientCodec, error) {
	if p.persistent {
		return p.cachedCodec, nil
	}
	var codec rpc.ClientCodec
	if p.useTLS {
		codec = newHttpPostCodec(p.codecHandle, p.serverAddr, p.persistent)
	} else {
		var err error
		codec, err = newTCPCodec(p.codecHandle, p.serverAddr)
		if err != nil {
			return nil, err
		}
	}

	return codec, nil
}

func (p *ClientProxy) Call(serviceMethod string, args interface{}, reply interface{}) error {
	client, err := p.rpcClient()
	if err != nil {
		return err
	}
	defer p.releaseClient(client)
	err = client.Call(serviceMethod, args, reply)
	if p.shouldRecord {
		p.RecordedReqs = append(p.RecordedReqs, RecordedRequest{serviceMethod, args, err, reply})
	}
	return err
}

func (p *ClientProxy) rpcClient() (*rpc.Client, error) {
	if p.persistent {
		return p.cachedClient, nil
	}

	codec, err := p.codec()
	if err != nil {
		return nil, err
	}
	return rpc.NewClientWithCodec(codec), nil
}

func (p *ClientProxy) releaseClient(client *rpc.Client) error {
	if !p.persistent {
		return client.Close()
	}
	return nil
}

func (p *ClientProxy) Close() {
	if p.persistent {
		p.cachedClient.Close()
		p.cachedCodec.Close()
	}
}

func (p *ClientProxy) StartRecording() {
	p.shouldRecord = true
	p.RecordedReqs = make([]RecordedRequest, 0)
}

func (p *ClientProxy) StopRecording() []RecordedRequest {
	p.shouldRecord = false
	return p.RecordedReqs
}

func newTCPCodec(codecHandle codec.Handle, serverAddr string) (rpc.ClientCodec, error) {
	var err error
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	return codec.GoRpc.ClientCodec(conn, codecHandle), nil
}
