package main

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"

	b "github.com/dimakogan/boosted-pir"
)

func main() {
	var db = []string{"A", "B", "C", "D", "E", "F"}
	randSource := rand.New(rand.NewSource(12345))
	stub := b.NewPIRServerStub(db, 100000, 1000, randSource)
	// Publish the receivers methods
	server := rpc.NewServer()
	err := server.RegisterName("PIRServer", stub)
	if err != nil {
		log.Fatal("Format of service PIRServer isn't correct. ", err)
	}
	// Listen to TPC connections on port 1234
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
	log.Printf("Serving RPC server on port %d", 1234)
	// Start accept incoming HTTP connections
	server.Accept(listener)
}
