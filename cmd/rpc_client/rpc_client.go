package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

	b "github.com/dimakogan/boosted-pir"

	"log"
	"net/rpc"
)

func main() {
	args := os.Args
	if len(args) < 4 {
		panic(fmt.Sprintf("Usage: %s <NUM-DB-RECORDS> <DB-RECORD-LEN-BYTES> <INDEX-TO-READ>", args[0]))
	}
	numRecords, err := strconv.Atoi(args[1])
	if err != nil {
		panic(fmt.Sprintf("Invalid NUM-DB-RECORDS: %s", args[1]))
	}
	setSize := int(math.Round(math.Sqrt(float64(numRecords))))
	nHints := setSize * int(math.Round(math.Log2(float64(numRecords))))
	pir := b.NewPirClientPunc(b.RandSource(), numRecords, nHints)

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
