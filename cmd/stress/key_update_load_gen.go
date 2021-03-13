package main

import (
	"fmt"
	"log"
	"math"

	. "checklist/driver"
	"checklist/updatable"
)

type keyUpdateLoadGen struct {
	numRows, updateSize int
}

func initKeyUpdateLoadGen(config *Config) *keyUpdateLoadGen {
	fmt.Printf("Connecting to %s (TLS: %t)...", config.ServerAddr, config.UseTLS)
	proxy, err := NewRpcProxy(config.ServerAddr, config.UseTLS, config.UsePersistent)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	defer proxy.Close()
	fmt.Printf("[OK]\n")

	fmt.Printf("Setting up remote DB...")
	if err := proxy.Configure(config.TestConfig, nil); err != nil {
		log.Fatalf("Failed to Configure: %s\n", err)
	}
	var numRows int
	if err = proxy.NumRows(0, &numRows); err != nil || numRows < config.NumRows*99/100 {
		log.Fatalf("Invalid number of rows on server: %d", numRows)
	}

	fmt.Printf("[OK] (num rows: %d)\n", config.NumRows)

	return &keyUpdateLoadGen{config.NumRows, config.UpdateSize}
}

func (s *keyUpdateLoadGen) request(proxy *RpcProxy) error {
	keyReq := updatable.KeyUpdatesReq{
		DefragTimestamp: math.MaxInt32,
		NextTimestamp:   int32(s.numRows - s.updateSize),
	}
	var keyResp updatable.KeyUpdatesResp
	err := proxy.KeyUpdates(keyReq, &keyResp)
	if err != nil {
		return fmt.Errorf("Failed to replay key update request %v, %s", keyReq, err)
	}
	if len(keyResp.Keys) != s.updateSize && int(keyResp.KeysRice.NumEntries)+1 != s.updateSize {
		return fmt.Errorf("Invalid size of key update, expected: %d, got: %d", s.updateSize, len(keyResp.Keys))
	}
	return nil
}

func (gen *keyUpdateLoadGen) debugStr() string {
	return ""
}

func (gen *keyUpdateLoadGen) reqRate() int {
	return 1
}
