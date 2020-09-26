package boosted

import (
	"fmt"
	"math"
	"testing"
)

func TestGGMPuncSetGen(t *testing.T) {
	gen := NewGGMSetGenerator(RandSource())

	tests := []struct {
		UnivSize int
		setSize  int
	}{
		{16, 5},
		{256, 16},
		{1 << 16, 10},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v", pair.UnivSize, pair.setSize),
			func(t *testing.T) {
				testPuncSetGenOnce(t, gen, pair.UnivSize, pair.setSize)
			})
	}
}

func TestGGMPuncSetPunc(t *testing.T) {
	gen := NewGGMSetGenerator(RandSource())

	tests := []struct {
		UnivSize int
		setSize  int
	}{
		{16, 5},
		{256, 16},
		{1 << 16, 10},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v", pair.UnivSize, pair.setSize),
			func(t *testing.T) {
				testPuncSetPunc(t, gen, pair.UnivSize, pair.setSize)
			})
	}
}

func TestGGMPuncSetGenWithPunc(t *testing.T) {
	gen := NewSetGenerator(NewGGMSetGenerator, MasterKey())

	tests := []struct {
		UnivSize int
		setSize  int
		with     int
	}{
		{16, 5, 0},
		{256, 16, 8},
		{1 << 16, 10, 7},
	}

	for _, pair := range tests {
		t.Run(fmt.Sprintf("%v %v %v", pair.UnivSize, pair.setSize, pair.with),
			func(t *testing.T) {
				testPuncSetGenWithPunc(t, gen, pair.UnivSize, pair.setSize, pair.with)
			})
	}
}

func TestGGMInSet(t *testing.T) {
	testInSet(t, NewGGMSetGenerator(RandSource()))
}

// Starting result:
// BenchmarkGGMEval-4   	    2926	    373090 ns/op	  104448 B/op	    2699 allocs/op
//
// Combine SetGenAndEval:
// BenchmarkGGMEval-4   	    3864	    294564 ns/op	   79615 B/op	    1663 allocs/op
//
// Preallocate keys for treeEvalAll:
// BenchmarkGGMEval-4   	    5881	    203453 ns/op	   53334 B/op	      14 allocs/op
func BenchmarkGGMEval(b *testing.B) {
	gen := NewSetGenerator(NewGGMSetGenerator, MasterKey())
	univSize := 1 * 1000 * 1000
	setSize := int(math.Sqrt(float64(univSize)))
	for i := 0; i < b.N; i++ {
		gen.SetGenAndEval(univSize, setSize)
	}
}
