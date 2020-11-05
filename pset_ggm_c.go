package boosted

/*
#cgo LDFLAGS: -L${SRCDIR}/prfggm-cpp -lpsetggm -lstdc++
#include "prfggm-cpp/pset_ggm.h"
*/
import "C"

//
// uint32_t univ_size, uint32_t set_size, const uint8_t* seed, uint32_t* out)

func PSetGGMEval(univSize, setSize uint32, seed []byte) []uint32 {
	out := make([]uint32, setSize)
	C.pset_ggm_eval(C.uint(univSize), C.uint(setSize), (*C.uchar)(&seed[0]), (*C.uint)(&out[0]))
	return out
}
