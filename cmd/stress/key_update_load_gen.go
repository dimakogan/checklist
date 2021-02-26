package main

import (
	"fmt"
	"math"

	. "github.com/dimakogan/boosted-pir"
)

type keyUpdateLoadGen struct {
	numRows, updateSize int
}

func initKeyUpdateLoadGen(config *Config, proxy *PirRpcProxy) *keyUpdateLoadGen {
	return &keyUpdateLoadGen{config.NumRows, config.UpdateSize}
}

func (s *keyUpdateLoadGen) request(proxy *PirRpcProxy) error {
	keyReq := KeyUpdatesReq{
		DefragTimestamp: math.MaxInt32,
		NextTimestamp:   int32(s.numRows - s.updateSize),
	}
	var keyResp KeyUpdatesResp
	err := proxy.KeyUpdates(keyReq, &keyResp)
	if err != nil {
		return fmt.Errorf("Failed to replay key update request %v, %s", keyReq, err)
	}
	if len(keyResp.Keys) != s.updateSize && int(keyResp.KeysRice.NumEntries)+1 != s.updateSize {
		return fmt.Errorf("Invalid size of key update, expected: %d, got: %d", s.updateSize, len(keyResp.Keys))
	}
	return nil
}
