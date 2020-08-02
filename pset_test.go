package boosted

import (
	"fmt"
	"testing"

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

func testPuncSetGenOnce(t *testing.T, gen SetGenerator, univSize int, setSize int) {
	key := gen.SetGen(univSize, setSize)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)
}

func TestPuncSetGen(t *testing.T) {
	gen := NewPrpSetGenerator(RandSource())
	tests := []struct {
		UnivSize int
		setSize  int
	}{
		{16, 5},
		{16, 16},
		{1 << 16, 10},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v", pair.UnivSize, pair.setSize),
			func(t *testing.T) {
				testPuncSetGenOnce(t, gen, pair.UnivSize, pair.setSize)
			})
	}
}

func testPuncSetGenWith(t *testing.T, gen SetGenerator, univSize int, setSize int, with int) {
	key := SetGenWith(gen, RandSource(), univSize, setSize, with)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)

	inSet := false
	for _, v := range set {
		inSet = inSet || (with == v)
	}

	assert.Assert(t, inSet)
}

func TestPuncSetGenWith(t *testing.T) {
	gen := NewPrpSetGenerator(RandSource())
	tests := []struct {
		UnivSize int
		setSize  int
		with     int
	}{
		{16, 5, 0},
		{256, 256, 8},
		{1 << 16, 10, 7},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v %v", pair.UnivSize, pair.setSize, pair.with),
			func(t *testing.T) {
				testPuncSetGenWith(t, gen, pair.UnivSize, pair.setSize, pair.with)
			})
	}
}

func testPuncSetPunc(t *testing.T, gen SetGenerator, univSize int, setSize int) {
	key := gen.SetGen(univSize, setSize)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)

	for i := 0; i < set.Size(); i++ {
		hole := set[i]
		pset := key.Punc(hole)
		assert.Equal(t, pset.Size(), setSize-1)

		inSet := false
		for _, v := range pset.Eval() {
			inSet = inSet || (hole == v)
			assert.Assert(t, key.InSet(v), "Element %d in punctured set %v but not in original set %v", v, pset.Eval(), set)
		}

		assert.Assert(t, !inSet)
	}
}

func testPuncSetGenWithPunc(t *testing.T, gen SetGenerator, univSize int, setSize int, with int) {
	key := SetGenWith(gen, RandSource(), univSize, setSize, with)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)

	inSet := false
	for _, v := range set {
		inSet = inSet || (with == v)
	}
	assert.Assert(t, inSet)

	pset := key.Punc(with)
	assert.Equal(t, pset.Size(), setSize-1)

	inSet = false
	for _, v := range pset.Eval() {
		//		assert.Equal(t, v, pset.ElemAt(i))
		inSet = inSet || (with == v)
		assert.Assert(t, key.InSet(v), "Element %d in punctured set %v, %v but not in original set %v", v, pset.Eval(), pset.Eval(), set)
	}

	assert.Assert(t, !inSet)
}

func TestPuncSetGenWithPunc(t *testing.T) {
	gen := NewPrpSetGenerator(RandSource())

	tests := []struct {
		UnivSize int
		setSize  int
		with     int
	}{
		{16, 5, 0},
		{16, 16, 8},
		{1 << 16, 10, 7},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v %v", pair.UnivSize, pair.setSize, pair.with),
			func(t *testing.T) {
				testPuncSetGenWithPunc(t, gen, pair.UnivSize, pair.setSize, pair.with)
			})
	}
}

func getElement(set Set) int {
	for _, k := range set {
		return k
	}

	return 0
}

func TestPRPInSet(t *testing.T) {
	testInSet(t, NewPrpSetGenerator(RandSource()))
}

func testInSet(t *testing.T, gen SetGenerator) {
	univSize := 1 << 4
	setSize := 4
	key := gen.SetGen(univSize, setSize)
	set := key.Eval()
	setHash := make(map[int]bool)
	for _, elem := range set {
		setHash[elem] = true
	}
	for i := 0; i < univSize; i++ {
		if _, exists := setHash[i]; exists {
			assert.Assert(t, key.InSet(i), "%v should have contained %d", set, i)
		} else {
			assert.Assert(t, !key.InSet(i), "%v should not contain %d", set, i)
		}
	}
}
