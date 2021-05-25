package psetggm

/*
#cgo amd64 CXXFLAGS: -msse2 -msse -march=native -maes -Ofast -std=c++11
#cgo arm64 CXXFLAGS: -march=armv8-a+fp+simd+crypto+crc -Ofast -std=c++11
#cgo LDFLAGS: -static-libstdc++
#include "pset_ggm.h"
#include "xor.h"
#include "answer.h"
*/
import "C"
import (
	"unsafe"
)

type GGMSetGeneratorC struct {
	workspace []byte
	cgen      *C.generator
}

func NewGGMSetGeneratorC(univSize, setSize int) *GGMSetGeneratorC {
	size := C.workspace_size(C.uint(univSize), C.uint(setSize))
	gen := GGMSetGeneratorC{
		workspace: make([]byte, size),
	}
	gen.cgen = C.pset_ggm_init(C.uint(univSize), C.uint(setSize),
		(*C.uchar)(&gen.workspace[0]))
	return &gen
}

func (gen *GGMSetGeneratorC) Eval(seed []byte, elems []int) {
	C.pset_ggm_eval(gen.cgen, (*C.uchar)(&seed[0]), (*C.ulonglong)(unsafe.Pointer(&elems[0])))
}

func (gen *GGMSetGeneratorC) Punc(seed []byte, pos int) []byte {
	pset := make([]byte, C.pset_buffer_size(gen.cgen))
	C.pset_ggm_punc(gen.cgen, (*C.uchar)(&seed[0]), C.uint(pos), (*C.uchar)(&pset[0]))
	return pset
}

func (gen *GGMSetGeneratorC) EvalPunctured(pset []byte, hole int, elems []int) {
	C.pset_ggm_eval_punc(gen.cgen, (*C.uchar)(&pset[0]), C.uint(hole), (*C.ulonglong)(unsafe.Pointer(&elems[0])))
}

func XorBlocks(db []byte, offsets []int, out []byte) {
	C.xor_rows((*C.uchar)(&db[0]), C.uint(len(db)), (*C.ulonglong)(unsafe.Pointer(&offsets[0])), C.uint(len(offsets)), C.uint(len(out)), (*C.uchar)(&out[0]))
}

func XorHashesByBitVector(db []byte, indexing []byte, out []byte) {
	C.xor_hashes_by_bit_vector((*C.uchar)(&db[0]), C.uint(len(db)),
		(*C.uchar)(&indexing[0]), (*C.uchar)(&out[0]))
}

func (gen *GGMSetGeneratorC) Distinct(elems []int) bool {
	return (C.distinct(gen.cgen, (*C.ulonglong)(unsafe.Pointer(&elems[0])), C.uint(len(elems))) != 0)
}

func FastAnswer(pset []byte, hole, univSize, setSize, shift int, db []byte, rowLen int, out []byte) {
	C.answer((*C.uchar)(&pset[0]), C.uint(hole), C.uint(univSize), C.uint(setSize), C.uint(shift),
		(*C.uchar)(&db[0]), C.uint(len(db)), C.uint(rowLen), C.uint(len(out)), (*C.uchar)(&out[0]))
}
