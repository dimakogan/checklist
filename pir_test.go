package boosted

import (
	"math/rand"
	"sync"
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

func BenchmarkNonUpdatable(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, config := range testConfigs() {
		db := MakeDB(randSource, config.NumRows, config.RowLen)

		server := NewPirServerByType(config.PirType, randSource, db)
		var mutex sync.Mutex
		leftServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      "Left/" + config.String(),
			mutex:     &mutex,
		}

		rightServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      "Right/" + config.String(),
			mutex:     &mutex,
		}

		client := NewPIRClient(NewPirClientByType(config.PirType,
			randSource), randSource,
			[2]PirServer{&leftServer, &rightServer})

		err := client.Init()
		assert.NilError(b, err)

		readIndex := rand.Intn(len(db))

		val, err := client.Read(readIndex)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[readIndex])
	}
}

func BenchmarkReadClient(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, config := range testConfigs() {
		db := MakeDB(randSource, config.NumRows, config.RowLen)
		server := NewPirServerByType(config.PirType, randSource, db)

		var mutex sync.Mutex
		pauseServer := pauseTimingServer{
			PirServer: server,
			mutex:     &mutex,
		}

		client := NewPIRClient(NewPirClientByType(config.PirType, randSource),
			randSource,
			[2]PirServer{&pauseServer, &pauseServer})

		assert.NilError(b, client.Init())

		b.Run(
			config.String(),
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
