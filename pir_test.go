package boosted

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"gotest.tools/assert"
)

var config *Config

func TestMain(m *testing.M) {
	config = new(Config).AddPirFlags().AddClientFlags().Parse()
	os.Exit(m.Run())
}
func TestPIRPunc(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)

	leftServer := NewPirServerPunc(&db)
	rightServer := NewPirServerPunc(&db)
	servers := [2]PirServer{leftServer, rightServer}
	client := NewPIRReader(RandSource(), servers)

	assert.NilError(t, client.Init(Punc))
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(readIndex))

	// Test refreshing by reading the same item again
	val, err = client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(readIndex))
}

func TestPunc(t *testing.T) {
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	presetRow := make(Row, 32)
	RandSource().Read(presetRow)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    100000,
		RowLen:     32,
		PresetRows: []RowIndexVal{{7, 0x7, presetRow}},
		PirType:    Punc,
		Updatable:  false,
	}, nil))

	client := NewPIRReader(RandSource(), [2]PirServer{driver, driver})

	err = client.Init(Punc)
	assert.NilError(t, err)

	// runtime.GC()
	// if memProf, err := os.Create("mem.prof"); err != nil {
	// 	panic(err)
	// } else {
	// 	pprof.WriteHeapProfile(memProf)
	// 	memProf.Close()
	// }
	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)
}

func TestPuncWithBlock(t *testing.T) {
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	presetRow := make(Row, 32)
	RandSource().Read(presetRow)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    100000,
		RowLen:     32,
		PresetRows: []RowIndexVal{{7, 0x7, presetRow}},
		PirType:    Punc,
		Updatable:  false,
	}, nil))

	client := NewPIRReader(RandSource(), [2]PirServer{driver, driver})

	err = client.Init(Punc)
	assert.NilError(t, err)

	// runtime.GC()
	// if memProf, err := os.Create("mem.prof"); err != nil {
	// 	panic(err)
	// } else {
	// 	pprof.WriteHeapProfile(memProf)
	// 	memProf.Close()
	// }
	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, presetRow)
}

func TestMatrix(t *testing.T) {
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    10000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x7, Row{'C', 'o', 'o', 'l'}}},
		PirType:    Matrix,
		Updatable:  false,
	}, nil))

	client := NewPIRReader(RandSource(), [2]PirServer{driver, driver})

	err = client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func TestDPF(t *testing.T) {
	driver, err := config.ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    512,
		RowLen:     32,
		PresetRows: []RowIndexVal{{128, 128, Row("12345678901234567890123456789012")}},
		PirType:    DPF,
		Updatable:  false,
	}, nil))

	client := NewPIRReader(RandSource(), [2]PirServer{driver, driver})

	err = client.Init(DPF)
	assert.NilError(t, err)

	val, err := client.Read(128)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("12345678901234567890123456789012"))
}

// Not testing this for now since disabled it
func DontTestPIRPuncKrzysztofTrick(t *testing.T) {
	src := RandSource()
	db := MakeDB(src, 4, 100)

	server := NewPirServerPunc(&db)

	for i := 0; i < 100; i++ {
		client := NewPIRReader(src, [2]PirServer{server, server})

		assert.NilError(t, client.Init(Punc))
		const readIndex = 2
		val, err := client.Read(readIndex)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, db.Row(readIndex))
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

func _TestMessageSizes(t *testing.T) {
	numRows := 3000000
	db := MakeDB(RandSource(), numRows, 32)

	server := NewPirServerPunc(&db)

	client := NewPirClientPunc(RandSource())
	clientWrapper := NewPIRReader(RandSource(), [2]PirServer{server, server})

	assert.NilError(t, clientWrapper.Init(Punc))

	qs, _ := client.Query(7)

	size, err := SerializedSizeOf(qs[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "query size: %d\n", size)
}

func _TestMessageSizesUpdatable(t *testing.T) {
	numRows := 4000000
	initialSize := 3000000
	keys := MakeKeys(RandSource(), numRows)
	db := MakeRows(RandSource(), numRows, 32)

	server := pirUpdatableServer{}
	server.AddRows(keys[0:initialSize], db[0:initialSize])

	client := NewPirClientUpdatable(RandSource(), Punc, [2]PirUpdatableServer{&server, &server})

	assert.NilError(t, client.Init())

	for i := 0; i < 50; i++ {
		server.AddRows(keys[initialSize+i*200:initialSize+(i+1)*200], db[initialSize+i*200:initialSize+(i+1)*200])
		assert.NilError(t, client.Update())

		qs, _ := client.waterfall.Query(int(keys[7]))
		size, err := SerializedSizeOf(qs[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stdout, "updatable request size: %d\n", size)
	}
}
