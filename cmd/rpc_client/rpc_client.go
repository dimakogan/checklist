package main

import (
	"fmt"
	"os"
	"strconv"

	b "github.com/dimakogan/boosted-pir"

	"log"
	"net/rpc"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		panic(fmt.Sprintf("Usage: %s <NUM-DB-RECORDS> <INDEX-TO-READ>", args[0]))
	}
	numRecords, err := strconv.Atoi(args[1])
	if err != nil {
		panic(fmt.Sprintf("Invalid NUM-DB-RECORDS: %s", args[1]))
	}

	idx, err := strconv.Atoi(args[2])
	if err != nil {
		panic(fmt.Sprintf("Invalid INDEX-TO-READ: %s", args[2]))
	}

	// Create a TCP connection to localhost on port 12345
	remote, err := rpc.DialHTTP("tcp", "localhost:12345")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	proxyLeft := b.NewPirRpcProxy(remote)
	proxyRight := b.NewPirRpcProxy(remote)
	client := b.NewPirClientPunc(b.RandSource(), numRecords, [2]b.PirServer{proxyLeft, proxyRight})

	val, err := client.Read(idx)
	if err != nil {
		log.Fatalf("Failed to read index %d: %w", idx, err)
	}
	fmt.Printf("Got value: %s from server\n", val)
}
