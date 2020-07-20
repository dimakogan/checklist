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

	for v := range set {
		assert.Assert(t, v < univSize && v >= 0)
	}
}

func testPuncSetGenOnce(t *testing.T, univSize int, setSize int) {
	key := SetGen(RandSource(), univSize, setSize)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)
}

func TestPuncSetGen(t *testing.T) {
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
				testPuncSetGenOnce(t, pair.UnivSize, pair.setSize)
			})
	}
}

func testPuncSetGenWith(t *testing.T, univSize int, setSize int, with int) {
	key := SetGenWith(RandSource(), univSize, setSize, with)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)

	inSet := false
	for v := range set {
		inSet = inSet || (with == v)
	}

	assert.Assert(t, inSet)
}

func TestPuncSetGenWith(t *testing.T) {
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
				testPuncSetGenWith(t, pair.UnivSize, pair.setSize, pair.with)
			})
	}
}

func testPuncSetGenWithPunc(t *testing.T, univSize int, setSize int, with int) {
	key := SetGenWith(RandSource(), univSize, setSize, with)
	set := key.Eval()
	checkSet(t, set, univSize, setSize)

	inSet := false
	for _, v := range set {
		inSet = inSet || (with == v)
	}
	assert.Assert(t, inSet)

	pset := key.Punc(with)
	assert.Equal(t, len(pset), setSize-1)

	inSet = false
	for _, v := range pset {
		inSet = inSet || (with == v)
	}

	assert.Assert(t, !inSet)
}

func TestPuncSetGenWithPunc(t *testing.T) {
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
				testPuncSetGenWithPunc(t, pair.UnivSize, pair.setSize, pair.with)
			})
	}
}

func getElement(set Set) int {
	for _, k := range set {
		return k
	}

	return 0
}

func TestRandomMemberExcept(t *testing.T) {
	key := SetGen(RandSource(), 1<<16, 2)
	set := key.Eval()

	v1 := getElement(set)
	v2 := key.RandomMemberExcept(RandSource(), v1)
	assert.Assert(t, v1 != v2)

	for _, k := range set {
		assert.Assert(t, k == v1 || k == v2)
	}
}

func TestInSet(t *testing.T) {
	univSize := 1 << 4
	setSize := 4
	key := SetGen(RandSource(), univSize, setSize)
	set := key.Eval()
	setHash := make(map[int]bool)
	for _, elem := range set {
		setHash[elem] = true
	}
	for i := 0; i < univSize; i++ {
		if _, exists := setHash[i]; exists {
			assert.Check(t, key.InSet(i))
		} else {
			assert.Check(t, !key.InSet(i))
		}
	}
}
