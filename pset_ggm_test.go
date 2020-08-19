package boosted

import (
	"fmt"
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
