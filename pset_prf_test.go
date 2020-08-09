package boosted

import (
	"fmt"
	"testing"
)

func TestPRFPuncSetGen(t *testing.T) {
	gen := NewPRFSetGenerator(RandSource())

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

func TestPRFPuncSetPunc(t *testing.T) {
	gen := NewPRFSetGenerator(RandSource())

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

func TestPRFPuncSetGenWithPunc(t *testing.T) {
	gen := NewPRFSetGenerator(RandSource())

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

func TestPRFInSet(t *testing.T) {
	testInSet(t, NewPRFSetGenerator(RandSource()))
}
