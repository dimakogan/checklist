package psetggm

/*
#cgo CXXFLAGS: -msse2 -msse -march=native -maes -Ofast
#include "pset_ggm.h"
#include "xor.h"
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

func XorRows(db []byte, rowLen int, numRows int, elems []int, out []byte) {
	C.xor_rows((*C.uchar)(&db[0]), C.uint(rowLen), C.uint(numRows), (*C.ulonglong)(unsafe.Pointer(&elems[0])), C.uint(len(elems)), (*C.uchar)(&out[0]))
}
