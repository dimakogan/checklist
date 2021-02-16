package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/rocketlaunchr/https-go"
	"github.com/ugorji/go/codec"

	b "github.com/dimakogan/boosted-pir"
)

func main() {
	port := flag.Int("p", 12345, "Listening port")
	useTLS := flag.Bool("tls", false, "Should use TLS")
	flag.Parse()

	driver, err := b.NewPirServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	var conn io.Closer
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		conn.Close()
	}()

	server := rpc.NewServer()
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	if *useTLS {
		// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
		server.HandleHTTP("/", "/debug")

		log.Printf("Serving RPC server over HTTPS on port %d\n", *port)
		// Use self-signed certificate
		httpServer, _ := https.Server(fmt.Sprintf("%d", *port), https.GenerateOptions{Host: "checklist.app"})
		conn = httpServer
		err = httpServer.ListenAndServeTLS("", "")
		if err == http.ErrServerClosed {
			log.Println("Server shutdown")
		} else if err != nil {
			log.Fatal("Failed to http.Serve, %w", err)
		}
	} else {
		serveTCP(server, *port)
	}
}

func serveTCP(server *rpc.Server, port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen tcp: %v", err)
	}
	log.Printf("Serving RPC server over TCP on port %d\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("TCP Accept failed: %+v\n", err)
		}
		go server.ServeCodec(codec.GoRpc.ServerCodec(conn, b.CodecHandle()))
	}
}
