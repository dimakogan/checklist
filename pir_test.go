package boosted

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 100)

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

func TestPIRPuncErasure(t *testing.T) {
	db := MakeDB(256, 100)

	server, err := NewPirServerErasure(RandSource(), db, DEFAULT_CHUNK_SIZE)
	assert.NilError(t, err)
	client := NewPIRClient(
		NewPirClientErasure(RandSource(), DEFAULT_CHUNK_SIZE),
		RandSource(),
		[2]PirServer{server, server})
	assert.NilError(t, err)
	assert.NilError(t, client.Init())
	const readIndex = 5
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// Test refreshing
	val, err = client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

// Not testing this for now since disabled it
func DontTestPIRPuncKrzysztofTrick(t *testing.T) {
	db := MakeDB(4, 100)
	src := RandSource()

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

func dbDimensions() []DBDimensions {
	var dims []DBDimensions
	numDBRecords :=
		//[]int{2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576}
		[]int{
			//1 << 10, 1 << 12, 1 << 14, 1 << 16, 1 << 18,
			// 1 << 16,
			1000000,
		}

	dbRecordSize := []int{1000}
	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(1 * 1024 * 1024 * 1024)

	for _, n := range numDBRecords {
		for _, recSize := range dbRecordSize {
			if int64(n)*int64(recSize) > maxDBSizeBytes {
				continue
			}
			dims = append(dims, DBDimensions{NumRecords: n, RecordSize: recSize})
		}
	}
	return dims
}

var chunkSizes = []int{DEFAULT_CHUNK_SIZE}

type benchmarkServer struct {
	PirServer
	b    *testing.B
	name string

	// Keep mutex to avoid parallelizm between two "servers" in  tests
	mutex *sync.Mutex
}

func (s *benchmarkServer) Hint(req HintReq, resp *HintResp) error {
	s.b.Run(
		"Hint/"+s.name,
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := s.PirServer.Hint(req, resp)
				assert.NilError(b, err)
			}
		})
	return s.PirServer.Hint(req, resp)
}

func (s *benchmarkServer) Answer(q QueryReq, resp *QueryResp) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.b.Run(
		"Answer/"+s.name,
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := s.PirServer.Answer(q, resp)
				assert.NilError(b, err)
			}
		})
	return nil
}

func BenchmarkPirPunc(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)

		server := NewPirServerPunc(randSource, db)
		var mutex sync.Mutex
		leftServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      fmt.Sprintf("Left/n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:     &mutex,
		}

		rightServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      fmt.Sprintf("Right/n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:     &mutex,
		}

		client := NewPIRClient(
			NewPirClientPunc(randSource),
			randSource,
			[2]PirServer{&leftServer, &rightServer})

		err := client.Init()
		assert.NilError(b, err)

		readIndex := rand.Intn(len(db))

		val, err := client.Read(readIndex)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[readIndex])
	}
}

func runPirErasure(b *testing.B, dim DBDimensions, chunkSize int) {
	randSource := rand.New(rand.NewSource(12345))
	db := MakeDBWithDimensions(dim)

	server, err := NewPirServerErasure(randSource, db, chunkSize)
	assert.NilError(b, err)

	var mutex sync.Mutex
	leftServer := benchmarkServer{
		PirServer: server,
		b:         b,
		name:      fmt.Sprintf("Left/n=%d,B=%d,CS=%d", dim.NumRecords, dim.RecordSize, chunkSize),
		mutex:     &mutex,
	}

	rightServer := benchmarkServer{
		PirServer: server,
		b:         b,
		name:      fmt.Sprintf("Right/n=%d,B=%d,CS=%d", dim.NumRecords, dim.RecordSize, chunkSize),
		mutex:     &mutex,
	}

	client := NewPIRClient(
		NewPirClientErasure(randSource, chunkSize),
		randSource,
		[2]PirServer{&leftServer, &rightServer})
	err = client.Init()
	assert.NilError(b, err)

	val, err := client.Read(5)
	assert.NilError(b, err)
	assert.DeepEqual(b, val, db[5])

}

func BenchmarkPirErasure(b *testing.B) {
	for _, dim := range dbDimensions() {
		for _, cs := range chunkSizes {
			runPirErasure(b, dim, cs)
		}
	}
}

type pauseTimingServer struct {
	PirServer
	b *testing.B

	// Keep mutex to avoid parallelizm between two "servers" in  tests
	mutex *sync.Mutex
}

func (s *pauseTimingServer) Hint(req HintReq, resp *HintResp) error {
	err := s.PirServer.Hint(req, resp)
	return err
}

func (s *pauseTimingServer) Answer(q QueryReq, resp *QueryResp) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.b.StopTimer()
	var err error
	err = s.PirServer.Answer(q, resp)
	s.b.StartTimer()
	return err
}

func BenchmarkPirErasureClient(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		server, err := NewPirServerErasure(randSource, db, DEFAULT_CHUNK_SIZE)
		assert.NilError(b, err)

		var mutex sync.Mutex
		pauseServer := pauseTimingServer{
			PirServer: server,
			mutex:     &mutex,
		}

		client := NewPIRClient(
			NewPirClientErasure(randSource, DEFAULT_CHUNK_SIZE),
			randSource,
			[2]PirServer{&pauseServer, &pauseServer})
		err = client.Init()
		assert.NilError(b, client.Init())

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					pauseServer.b = b
					val, err := client.Read(5)
					assert.NilError(b, err)
					assert.DeepEqual(b, val, db[5])
				}
			})
	}
}

func BenchmarkHintOnce(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	server := NewPirServerPunc(randSource, db)
	benchmarkServer := benchmarkServer{
		PirServer: server,
		b:         b,
		name:      fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
	}

	client := NewPIRClient(
		NewPirClientPunc(randSource),
		randSource,
		[2]PirServer{&benchmarkServer, server})
	err := client.Init()
	assert.NilError(b, err)
}

func BenchmarkNothingRandom(b *testing.B) {
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	nHints := 1024
	setLen := 1024

	out := make(Row, dim.RecordSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				q := ((123124124 * k) + 912812367) % dim.NumRecords
				xorInto(out, db[q])
			}
		}
	}
}

func BenchmarkNothingLinear(b *testing.B) {
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	nHints := 1024
	setLen := 1024

	out := make(Row, dim.RecordSize)
	b.ResetTimer()
	q := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				xorInto(out, db[q])
				q = (q + 1) % dim.NumRecords
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
