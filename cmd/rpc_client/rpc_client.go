package main

import (
	"fmt"

	b "github.com/dimakogan/boosted-pir"

	"log"
	"net/rpc"
)

func main() {
	pir := b.NewPirClientStub()
	if pir == nil {
		log.Fatal("Failed to create PIRClient")
	}

	// Create a TCP connection to localhost on port 1234
	remote, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	client, err := b.NewRpcPirClient(remote, pir)
	if err != nil {
		log.Fatalf("Failed to create RPC PIR client: %w", err)
	}
	const readIndex = 2
	val, err := client.Read(readIndex)
	if err != nil {
		log.Fatalf("Failed to read index %d: %w", readIndex, err)
	}
	fmt.Printf("Got value: %s from server\n", val)
}
