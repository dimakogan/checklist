package boosted

import (
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func TestPIRPunc(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)

	leftServer := NewPirServerPunc(RandSource(), db)
	rightServer := NewPirServerPunc(RandSource(), db)
	servers := [2]PirServer{leftServer, rightServer}
	client := NewPIRClient(
		NewPirClientPunc(RandSource()),
		RandSource(),
		servers)

	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// Test refreshing by reading the same item again
	val, err = client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPunc(t *testing.T) {
	driver, err := ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    10000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x7, Row{'C', 'o', 'o', 'l'}}},
		PirType:    Punc,
		Updatable:  false,
	}, nil))

	//client, err := NewPirClientErasure(RandSource(), 1000, DEFAULT_CHUNK_SIZE, [2]PirServer{proxy, proxy})
	client := NewPIRClient(NewPirClientPunc(RandSource()), RandSource(), [2]PirServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func TestMatrix(t *testing.T) {
	driver, err := ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    10000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x7, Row{'C', 'o', 'o', 'l'}}},
		PirType:    Matrix,
		Updatable:  false,
	}, nil))

	//client, err := NewPirClientErasure(RandSource(), 1000, DEFAULT_CHUNK_SIZE, [2]PirServer{proxy, proxy})
	client := NewPIRClient(NewPirClientMatrix(RandSource()), RandSource(), [2]PirServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func TestDPF(t *testing.T) {
	driver, err := ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    512,
		RowLen:     32,
		PresetRows: []RowIndexVal{{128, 128, Row("12345678901234567890123456789012")}},
		PirType:    DPF,
		Updatable:  false,
	}, nil))

	//client, err := NewPirClientErasure(RandSource(), 1000, DEFAULT_CHUNK_SIZE, [2]PirServer{proxy, proxy})
	client := NewPIRClient(NewPIRDPFClient(RandSource()), RandSource(), [2]PirServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(128)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("12345678901234567890123456789012"))
}

// Not testing this for now since disabled it
func DontTestPIRPuncKrzysztofTrick(t *testing.T) {
	src := RandSource()
	db := MakeDB(src, 4, 100)

	server := NewPirServerPunc(src, db)

	for i := 0; i < 100; i++ {
		client := NewPIRClient(
			NewPirClientPunc(src),
			src,
			[2]PirServer{server, server})

		assert.NilError(t, client.Init())
		const readIndex = 2
		val, err := client.Read(readIndex)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, db[readIndex])
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[i%len(letterBytes)]
	}
	return string(b)
}

// Save results to avoid compiler optimizations.
var hint *HintResp
var resp *QueryResp

var chunkSizes = []int{DEFAULT_CHUNK_SIZE}

func BenchmarkNothingRandom(b *testing.B) {
	config := TestConfig{NumRows: 1024 * 1024, RowLen: 1024}
	db := MakeDB(RandSource(), config.NumRows, config.RowLen)

	nHints := 1024
	setLen := 1024

	out := make(Row, config.RowLen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				q := ((123124124 * k) + 912812367) % config.NumRows
				xorInto(out, db[q])
			}
		}
	}
}

func BenchmarkNothingLinear(b *testing.B) {
	config := TestConfig{NumRows: 1024 * 1024, RowLen: 1024}
	db := MakeDB(RandSource(), config.NumRows, config.RowLen)

	nHints := 1024
	setLen := 1024

	out := make(Row, config.RowLen)
	b.ResetTimer()
	q := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				xorInto(out, db[q])
				q = (q + 1) % config.NumRows
			}
		}
	}
}

func TestSample(t *testing.T) {
	client := NewPirClientPunc(RandSource())
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

// func TestDPFEval(t *testing.T) {
// 	key1, key2 := dpf.Gen(128, 9)
// 	log.Printf("\n%v\n%v\n", dpf.EvalFull(key1, 9)[0:64], dpf.EvalFull(key2, 9))

// 	for i := 0; i < 1<<9; i++ {
// 		assert.Assert(t, dpf.Eval(key1, uint64(i), 9) != dpf.Eval(key2, uint64(i), 9) {
// 			log.Printf("differ at: %d\n", i)
// 		}
// 	}

// }

// func TestEvalFull(test *testing.T) {
// 	logN := uint64(9)
// 	alpha := uint64(128)
// 	a, b := dpf.Gen(alpha, logN)
// 	aa := dpf.EvalFull(a, logN)
// 	bb := dpf.EvalFull(b, logN)
// 	for i := uint64(0); i < (uint64(1) << logN); i++ {
// 		aaa := (aa[i/8] >> (i % 8)) & 1
// 		bbb := (bb[i/8] >> (i % 8)) & 1
// 		if (aaa^bbb == 1 && i != alpha) || (aaa^bbb == 0 && i == alpha) {
// 			test.Fail()
// 		}
// 	}
// }
