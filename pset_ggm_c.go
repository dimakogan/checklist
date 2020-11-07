package boosted

/*
#cgo LDFLAGS: -L${SRCDIR}/prfggm-cpp -lpsetggm -lstdc++
#include "prfggm-cpp/pset_ggm.h"
*/
import "C"
import (
	"encoding/binary"
	"math"
	"math/rand"
)

//
// uint32_t univ_size, uint32_t set_size, const uint8_t* seed, uint32_t* out)

func PSetGGMEval(univSize, setSize int, seed []byte) []int {
	// out := make([]int, setSize)
	// C.pset_ggm_eval(C.uint(univSize), C.uint(setSize), (*C.uchar)(&seed[0]), (*C.uint)(&out[0]))
	src := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed))))
	gen := NewGGMSetGenerator(src)
	_, elems := gen.SetGenAndEval(univSize, setSize)
	return elems
}

func PSetGGMPunc(univSize, setSize int, seed []byte, pos int) [][]byte {
	src := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed))))
	gen := NewGGMSetGenerator(src)
	set, elems := gen.SetGenAndEval(univSize, setSize)
	pset := set.Punc(elems[pos])
	return pset.Keys
}

func PSetGGMEvalPunctured(univSize, setSize int, keys [][]byte, hole int) []int {
	height := int(math.Ceil(math.Log2(float64(setSize + 1))))

	pset := puncturedGGMSet{Keys: keys, Hole: hole, SetSize: setSize, UnivSize: univSize, Height: height}
	return pset.Eval()
}
