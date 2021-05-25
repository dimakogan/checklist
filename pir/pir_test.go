package pir

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 100)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Punc)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))

	// Test refreshing by reading the same item again
	val, err = client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))

}

func TestMatrix(t *testing.T) {
	db := MakeDB(10000, 4)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))
}

func TestDPF(t *testing.T) {
	db := MakeDB(512, 32)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(DPF)
	assert.NilError(t, err)

	val, err := client.Read(128)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(128))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[i%len(letterBytes)]
	}
	return string(b)
}

func TestSample(t *testing.T) {
	client := puncClient{randSource: RandSource()}
	assert.Equal(t, 1, client.sample(10, 0, 10))
	assert.Equal(t, 2, client.sample(0, 10, 10))
	assert.Equal(t, 0, client.sample(0, 0, 10))
	count := make([]int, 3)
	for i := 0; i < 1000; i++ {
		count[client.sample(10, 10, 30)]++
	}
	for _, c := range count {
		assert.Check(t, c < 380)
		assert.Check(t, c > 280)
	}
}
