package main

import (
	"fmt"

	b "github.com/dimakogan/boosted-pir"

	"log"
	"net/rpc"
)

func main() {
	client := b.NewPirClientStub()
	if client == nil {
		log.Fatal("Failed to create PIRClient")
	}

	// Create a TCP connection to localhost on port 1234
	remote, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	runClient(remote, client)
}

func runClient(remote *rpc.Client, client b.PIRClient) {
	hintReq, err := client.RequestHint()
	if err != nil {
		log.Fatal("Client failed to RequestHint", err)
	}
	var hintResp b.HintResp
	if err = remote.Call("PIRServer.Hint", hintReq, &hintResp); err != nil {
		log.Fatal("Remote PIRServer.Hint failed, ", err)
	}
	if err = client.InitHint(&hintResp); err != nil {
		log.Fatal("Client failed to InitHint, ", err)
	}

	const readIndex = 2
	queryReq, err := client.Query(readIndex)
	if err != nil {
		log.Fatal("Client failed to Query, ", err)
	}

	var queryResp b.QueryResp
	if err = remote.Call("PIRServer.Answer", queryReq[0], &queryResp); err != nil {
		log.Fatal("Remote Answer failed, ", err)
	}

	val, err := client.Reconstruct([]*b.QueryResp{&queryResp})
	if err != nil {
		log.Fatal("Client failed to Reconstruct, ", err)
	}

	fmt.Printf("Got value: %s from server\n", val)
}
