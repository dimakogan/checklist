package rpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
	"strings"

	"github.com/rocketlaunchr/https-go"
	"github.com/ugorji/go/codec"
)

type httpServerCodec struct {
	httpResponse http.ResponseWriter

	encoder *codec.Encoder
	decoder *codec.Decoder
}

func (c *httpServerCodec) WriteResponse(header *rpc.Response, body interface{}) error {
	if header.Error != "" {
		c.httpResponse.Header().Set("Go-Error", header.Error)
		c.httpResponse.WriteHeader(http.StatusInternalServerError)
	}
	err := c.encoder.Encode(header)
	if err != nil {
		return err
	}
	return c.encoder.Encode(body)
}

func (c *httpServerCodec) Close() error {
	return nil
}

func (c *httpServerCodec) ReadRequestHeader(header *rpc.Request) error {
	err := c.decoder.Decode(header)
	return err
}

func (c *httpServerCodec) ReadRequestBody(body interface{}) error {
	err := c.decoder.Decode(body)
	return err
}

type Server interface {
	RegisterName(name string, rcvr interface{}) error
	Serve() error
	Close() error
}

type httpRpcServer struct {
	io.Closer
	httpServer *http.Server
	*rpc.Server
}

func (s *httpRpcServer) Serve() error {
	log.Printf("Serving RPC server over HTTPS on %s\n", s.httpServer.Addr)
	err := s.httpServer.ListenAndServeTLS("", "")
	if err == http.ErrServerClosed {
		log.Println("Server shutdown")
		return nil
	}
	return err
}

func NewServer(port int, useTLS bool, types []reflect.Type) (Server, error) {
	rpcServer := rpc.NewServer()

	codecHandle := CodecHandle(types)

	if useTLS {
		httpSrv, err := https.Server(fmt.Sprintf("%d", port),
			https.GenerateOptions{Host: "test.app", ECDSACurve: "P256"})
		if err != nil {
			return nil, err
		}
		server := httpRpcServer{httpSrv, httpSrv, rpcServer}
		httpSrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, rpc.DefaultRPCPath) {
				w.Header().Set("Content-type", "application/octet-stream")
				codec := httpServerCodec{
					httpResponse: w,
					encoder:      codec.NewEncoder(w, codecHandle),
					decoder:      codec.NewDecoder(r.Body, codecHandle)}
				err := server.Server.ServeRequest(&codec)
				if err != nil {
					w.Header().Set("Go-Error", fmt.Sprintf("%s", err))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		})
		return &server, nil
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("Failed to listen tcp: %v", err)
	}
	return &tcpRpcServer{ln, rpcServer, codecHandle}, nil
}

type tcpRpcServer struct {
	net.Listener
	*rpc.Server

	codecHandle codec.Handle
}

func (s *tcpRpcServer) Serve() error {
	log.Printf("Serving RPC server over TCP on %s\n", s.Addr().String())
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return fmt.Errorf("TCP Accept failed: %+v\n", err)
		}
		go s.Server.ServeCodec(codec.GoRpc.ServerCodec(conn, s.codecHandle))
	}
}
