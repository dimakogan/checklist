package pir

import (
	"flag"
	"fmt"
	"math"
	"testing"

	"checklist/psetggm"

	"gotest.tools/assert"
)

func TestMathMod(t *testing.T) {
	assert.Assert(t, MathMod(5, 10) == 5)
	assert.Assert(t, MathMod(10, 10) == 0)
	assert.Assert(t, MathMod(-2, 10) == 8)
	assert.Assert(t, MathMod(-1, 10) == 9)
	assert.Assert(t, MathMod(-10, 10) == 0)
	assert.Assert(t, MathMod(20, 10) == 0)
	assert.Assert(t, MathMod(27, 10) == 7)
	assert.Assert(t, MathMod(-27, 10) == 3)
	assert.Assert(t, MathMod(-30, 10) == 0)
	assert.Assert(t, MathMod(-100, 10) == 0)
	assert.Assert(t, MathMod(-99, 10) == 1)
}

func checkSet(t *testing.T, set Set, univSize int, setSize int) {
	assert.Equal(t, len(set), setSize)

	for _, v := range set {
		assert.Assert(t, v < univSize && v >= 0)
	}
}

func testGenWith(t *testing.T, gen *SetGenerator, univSize int, setSize int, with int) {
	set := gen.GenWith(with)

	checkSet(t, set.elems, univSize, setSize)

	inSet := false
	for _, v := range set.elems {
		inSet = inSet || (with == v)
	}

	assert.Assert(t, inSet)
}

func testPunc(t *testing.T, gen SetGenerator, univSize int, setSize int) {
	var set PuncturableSet
	gen.Gen(&set)
	checkSet(t, set.elems, univSize, setSize)

	setHash := make(map[int]bool)
	for _, elem := range set.elems {
		setHash[elem] = true
	}

	for _, hole := range set.elems {
		pset := gen.Punc(set, hole)
		elems := pset.Eval()
		assert.Equal(t, len(elems), setSize-1)

		inSet := false
		for _, v := range elems {
			inSet = inSet || (hole == v)
			assert.Assert(t, setHash[v], "Element %d in punctured set %v but not in original set %v", v, pset.Eval(), set.elems)
		}

		assert.Assert(t, !inSet)
	}
}

func testGenWithPunc(t *testing.T, gen *SetGenerator, univSize int, setSize int, with int) {
	set := gen.GenWith(with)
	checkSet(t, set.elems, univSize, setSize)

	inSet := false
	for _, v := range set.elems {
		inSet = inSet || (with == v)
	}
	assert.Assert(t, inSet)

	setHash := make(map[int]bool)
	for _, elem := range set.elems {
		setHash[elem] = true
	}

	pset := gen.Punc(set, with)
	assert.Equal(t, pset.SetSize, setSize-1)

	inSet = false
	for _, v := range pset.Eval() {
		//		assert.Equal(t, v, pset.ElemAt(i))
		inSet = inSet || (with == v)
		assert.Assert(t, setHash[v], "Element %d in punctured set %v, %v but not in original set %v", v, pset.Eval(), set.elems)
	}

	assert.Assert(t, !inSet)
}

func getElement(set Set) int {
	for _, k := range set {
		return k
	}

	return 0
}

func TestPuncSetGen(t *testing.T) {
	tests := []struct {
		UnivSize int
		setSize  int
	}{
		{16, 5},
		{256, 16},
		{1 << 16, 10},
	}

	var set PuncturableSet
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v", test.UnivSize, test.setSize),
			func(t *testing.T) {
				gen := NewSetGenerator(masterKey, 0, test.UnivSize, test.setSize)
				gen.Gen(&set)
				checkSet(t, set.elems, test.UnivSize, test.setSize)
			})
	}
}

func TestPuncSetPunc(t *testing.T) {
	tests := []struct {
		UnivSize int
		setSize  int
	}{
		{16, 5},
		{256, 16},
		{1 << 16, 10},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v", test.UnivSize, test.setSize),
			func(t *testing.T) {
				gen := NewSetGenerator(masterKey, 0, test.UnivSize, test.setSize)
				testPunc(t, gen, test.UnivSize, test.setSize)
			})
	}
}

func TestPuncSetGenWith(t *testing.T) {
	tests := []struct {
		UnivSize int
		setSize  int
		with     int
	}{
		{16, 5, 0},
		{256, 16, 8},
		{1 << 16, 10, 7},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v %v", test.UnivSize, test.setSize, test.with),
			func(t *testing.T) {
				gen := NewSetGenerator(masterKey, 0, test.UnivSize, test.setSize)
				testGenWith(t, &gen, test.UnivSize, test.setSize, test.with)
			})
	}
}

func TestPuncSetGenWithPunc(t *testing.T) {
	tests := []struct {
		UnivSize int
		setSize  int
		with     int
	}{
		{16, 5, 0},
		{256, 16, 8},
		{1 << 16, 10, 7},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v %v", test.UnivSize, test.setSize, test.with),
			func(t *testing.T) {
				gen := NewSetGenerator(masterKey, 0, test.UnivSize, test.setSize)
				testGenWithPunc(t, &gen, test.UnivSize, test.setSize, test.with)
			})
	}
}

var univSize = flag.Int("univSize", 10000, "universe size for puncturable-set test")

// Starting result:
// BenchmarkGGMEval-4   	    2926	    373090 ns/op	  104448 B/op	    2699 allocs/op
//
// Combine SetGenAndEval:
// BenchmarkGGMEval-4   	    3864	    294564 ns/op	   79615 B/op	    1663 allocs/op
//
// Preallocate keys for treeEvalAll:
// BenchmarkGGMEval-4   	    5881	    203453 ns/op	   53334 B/op	      14 allocs/op
func BenchmarkPuncSetGen(b *testing.B) {
	setSize := int(math.Sqrt(float64(*univSize)))
	gen := NewGGMSetGenerator(RandSource())
	b.Run(fmt.Sprintf("UnivSize=%d", *univSize), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.SetGenAndEval(*univSize, setSize)
		}
	})
}

func BenchmarkGGMEvalC(b *testing.B) {
	setSize := int(math.Sqrt(float64(*univSize)))

	gen := psetggm.NewGGMSetGeneratorC(*univSize, setSize)
	set := make([]int, setSize)
	seed := make([]byte, 16)
	b.Run(fmt.Sprintf("UnivSize=%d", *univSize), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			seed[0] = (byte)(i % 256)
			gen.Eval(seed, set)
		}
	})
}

func BenchmarkGen(b *testing.B) {
	setSize := int(math.Sqrt(float64(*univSize)))

	gen := NewSetGenerator(masterKey, 0, *univSize, setSize)
	var set PuncturableSet
	b.Run(fmt.Sprintf("UnivSize=%d", *univSize), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.gen(&set)
		}
	})
}
