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
	var db = make([]b.Row, b.CHUNK_SIZE)
	for i := 0; i < len(db); i++ {
		db[i] = b.Row{byte('A' + i), byte('A' + i), byte('A' + i)}
	}

	driver, err := b.NewPirRpcServer(db)
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	// Listen to TPC connections on port 12345
	listener, e := net.Listen("tcp", ":12345")
	if e != nil {
		log.Fatalf("Listen error: %s", e)
	}
	server := rpc.NewServer()
	if err := server.Register(driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")

	log.Printf("Serving RPC server on %s", ":12345")
	// Start accept incoming HTTP connections
	if e = http.Serve(listener, nil); e != nil {
		log.Fatal("Failed to http.Serve, %w", e)
	}
}
