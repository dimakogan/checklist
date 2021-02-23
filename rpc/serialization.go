package rpc

import "github.com/ugorji/go/codec"

func CodecHandle() codec.Handle {
	h := codec.BincHandle{}
	h.StructToArray = true
	h.OptimumSize = true
	return &h
}
