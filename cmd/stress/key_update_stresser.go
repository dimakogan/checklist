package main

import (
	"fmt"
	"math"

	. "github.com/dimakogan/boosted-pir"
)

type keyUpdateStresser struct {
	numRows, updateSize int
}

func initKeyUpdateStresser(config *Configurator, proxy *PirRpcProxy) *keyUpdateStresser {
	return &keyUpdateStresser{config.NumRows, config.UpdateSize}
}

func (s *keyUpdateStresser) request(proxy *PirRpcProxy) error {
	keyReq := KeyUpdatesReq{
		DefragTimestamp: math.MaxInt32,
		NextTimestamp:   int32(s.numRows - s.updateSize),
	}
	var keyResp KeyUpdatesResp
	err := proxy.KeyUpdates(keyReq, &keyResp)
	if err != nil {
		return fmt.Errorf("Failed to replay key update request %v, %s", keyReq, err)
	}
	if len(keyResp.Keys) != s.updateSize {
		return fmt.Errorf("Invalid size of key update, expected: %d, got: %d", s.numRows, len(keyResp.Keys))
	}
	return nil
}
