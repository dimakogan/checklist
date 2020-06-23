package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	b "github.com/dimakogan/boosted-pir"
)

func main() {
	// Some easy to test initial values.
	var db = []b.Row{{'A'}, {'B'}, {'C'}, {'D'}}

	driver := b.NewPirRpcServer(db)

	// Listen to TPC connections on port 1234
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error, %w ", e)
	}
	server := rpc.NewServer()
	if err := server.Register(driver); err != nil {
		log.Fatal("Failed to register PIRServer, %w", err)
	}

	// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")

	log.Printf("Serving RPC server on %s", ":1234")
	// Start accept incoming HTTP connections
	if e = http.Serve(listener, nil); e != nil {
		log.Fatal("Failed to http.Serve, %w", e)
	}
}
