package rpc

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

func CodecHandle(types []reflect.Type) codec.Handle {
	h := codec.BincHandle{}
	h.StructToArray = true
	h.OptimumSize = true
	h.PreferPointerForStructOrArray = true

	for i, t := range types {
		err := h.SetBytesExt(t, uint64(0x10+i), codec.SelfExt)
		if err != nil {
			panic(err)
		}
	}

	return &h
}
