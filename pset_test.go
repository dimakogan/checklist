package boosted

import (
	"math/rand"
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

func randSource() *rand.Rand {
  return rand.New(rand.NewSource(17))
}

func checkSet(t *testing.T, set Set, univSize int, setSize int) {
  assert.Equal(t, len(set), setSize)

  for v := range(set) {
    assert.Assert(t, v < univSize && v >= 0)
  }
}

func testPuncSetGenOnce(t *testing.T, univSize int, setSize int) {
  key := SetGen(randSource(), univSize, setSize)
  set := key.Eval()
  checkSet(t, set, univSize, setSize)
}

func TestPuncSetGen_10_5(t *testing.T) {
  testPuncSetGenOnce(t, 10, 5)
}

func TestPuncSetGen_10_10(t *testing.T) {
  testPuncSetGenOnce(t, 10, 10)
}

func TestPuncSetGen_1_1(t *testing.T) {
  testPuncSetGenOnce(t, 1, 1)
}

func TestPuncSetGen_100000_10(t *testing.T) {
  testPuncSetGenOnce(t, 100000, 10)
}

func testPuncSetGenWith(t *testing.T, univSize int, setSize int, with int) {
  key := SetGenWith(randSource(), univSize, setSize, with)
  set := key.Eval()
  checkSet(t, set, univSize, setSize)

  inSet := false
  for v := range(set) {
    inSet = inSet || (with == v)
  }

  assert.Assert(t, inSet)
}

func TestPuncSetGenWith_10_5(t *testing.T) {
  testPuncSetGenWith(t, 10, 5, 0)
}

func TestPuncSetGenWith_10_10(t *testing.T) {
  testPuncSetGenWith(t, 10, 10, 8)
}

func TestPuncSetGenWith_1_1(t *testing.T) {
  testPuncSetGenWith(t, 1, 1, 0)
}

func TestPuncSetGenWith_100000_10(t *testing.T) {
  testPuncSetGenWith(t, 100000, 10, 7)
}

func testPuncSetGenWithPunc(t *testing.T, univSize int, setSize int, with int) {
  key := SetGenWith(randSource(), univSize, setSize, with)
  set := key.Eval()
  checkSet(t, set, univSize, setSize)

  inSet := false
  for v := range(set) {
    inSet = inSet || (with == v)
  }
  assert.Assert(t, inSet)

  pkey := key.Punc(with)
  pset := pkey.Eval()
  assert.Equal(t, len(pset), setSize-1)

  inSet = false
  for v := range(pset) {
    inSet = inSet || (with == v)
  }

  assert.Assert(t, !inSet)
}


func TestPuncSetGenWithPunc_10_5(t *testing.T) {
  testPuncSetGenWithPunc(t, 10, 5, 0)
}

func TestPuncSetGenWithPunc_10_10(t *testing.T) {
  testPuncSetGenWithPunc(t, 10, 10, 8)
}

func TestPuncSetGenWithPunc_1_1(t *testing.T) {
  testPuncSetGenWithPunc(t, 1, 1, 0)
}

func TestPuncSetGenWithPunc_100000_10(t *testing.T) {
  testPuncSetGenWithPunc(t, 100000, 10, 7)
}


func TestPuncSetShift(t *testing.T) {
  key := SetGen(randSource(), 10, 5)
  set := key.Eval()

  key.Shift(1)
  set2 := key.Eval()

  for i := range set2 {
    j := MathMod(i - 1, key.univSize)
    assert.Assert(t, set[j])
  }
}

