package main

import (
	"fmt"
	"os"
	"strconv"

	b "github.com/dimakogan/boosted-pir"

	"log"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		panic(fmt.Sprintf("Usage: %s <INDEX-TO-READ>", args[0]))
	}

	idx, err := strconv.Atoi(args[1])
	if err != nil {
		panic(fmt.Sprintf("Invalid INDEX-TO-READ: %s", args[1]))
	}

	proxyLeft, err := b.NewPirRpcProxy("localhost:12345")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	proxyRight, err := b.NewPirRpcProxy("localhost:12345")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	client := b.NewPIRClient(
		b.NewPirClientPunc(b.RandSource()),
		b.RandSource(),
		[2]b.PirServer{proxyLeft, proxyRight})

	val, err := client.Read(idx)
	if err != nil {
		log.Fatalf("Failed to read index %d: %w", idx, err)
	}
	fmt.Printf("Got value: %s from server\n", val)
}
