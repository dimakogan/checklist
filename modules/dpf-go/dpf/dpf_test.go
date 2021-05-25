package dpf

import (
	"fmt"
	"testing"
)

func BenchmarkEvalFull(bench *testing.B) {
	logN := uint64(28)
	a, _ := Gen(0, logN)
	bench.ResetTimer()
	//fmt.Println("Ka: ", a)
	//fmt.Println("Kb: ", b)
	//for i:= uint64(0); i < (uint64(1) << logN); i++ {
	//	aa := dpf.Eval(a, i, logN)
	//	bb := dpf.Eval(b, i, logN)
	//	fmt.Println(i,"\t", aa,bb, aa^bb)
	//}
	for i := 0; i < bench.N; i++ {
		EvalFull(a, logN)
	}
}

func BenchmarkXor16(bench *testing.B) {
	a := new(block)
	b := new(block)
	c := new(block)
	for i := 0; i < bench.N; i++ {
		xor16(&c[0], &b[0], &a[0])
	}
}

func TestEval(test *testing.T) {
	logN := uint64(8)
	alpha := uint64(123)
	a, b := Gen(alpha, logN)
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		aa := Eval(a, i, logN)
		bb := Eval(b, i, logN)
		if (aa^bb == 1 && i != alpha) || (aa^bb == 0 && i == alpha) {
			test.Fail()
		}
	}
}

func TestEvalFull(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(128)
	a, b := Gen(alpha, logN)
	aa := EvalFull(a, logN)
	bb := EvalFull(b, logN)
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		aaa := (aa[i/8] >> (i % 8)) & 1
		bbb := (bb[i/8] >> (i % 8)) & 1
		if (aaa^bbb == 1 && i != alpha) || (aaa^bbb == 0 && i == alpha) {
			test.Fail()
		}
	}
}

func TestEvalFullShort(test *testing.T) {
	logN := uint64(3)
	alpha := uint64(1)
	a, b := Gen(alpha, logN)
	aa := EvalFull(a, logN)
	bb := EvalFull(b, logN)
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		aaa := (aa[i/8] >> (i % 8)) & 1
		bbb := (bb[i/8] >> (i % 8)) & 1
		if (aaa^bbb == 1 && i != alpha) || (aaa^bbb == 0 && i == alpha) {
			test.Fail()
		}
	}
}


func DebugAES(test *testing.T) { 
	var prfkeyL = []byte{36, 156, 50, 234, 92, 230, 49, 9, 174, 170, 205, 160, 98, 236, 29, 243}
	var prfkeyR = []byte{209, 12, 199, 173, 29, 74, 44, 128, 194, 224, 14, 44, 2, 201, 110, 28}
	var keyL = make([]uint32, 11*4)
	var keyR = make([]uint32, 11*4)	

	expandKeyAsm(&prfkeyL[0], &keyL[0])
	expandKeyAsm(&prfkeyR[0], &keyR[0])
	fmt.Printf("%+v\n", keyL)
	fmt.Printf("%+v\n", keyR)

	seed :=  []byte{123, 56, 5, 24, 9, 20, 4, 9, 14, 10, 25, 10, 9, 26, 29, 43}
	s0 := new(block)
	aes128MMO(&keyL[0], &s0[0], &seed[0])
	fmt.Printf("%+v\n", seed)
	fmt.Printf("%+v\n", *s0)

	xor16(&s0[0], &s0[0], &s0[0])
	fmt.Printf("%+v\n", s0)
}