package boosted

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

var chunkSizes = []int{DEFAULT_CHUNK_SIZE}

func TestPIRPuncErasure(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)

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

func runPirErasure(b *testing.B, config TestConfig, chunkSize int) {
	randSource := rand.New(rand.NewSource(12345))
	db := MakeDB(randSource, config.NumRows, config.RowLen)

	server, err := NewPirServerErasure(randSource, db, chunkSize)
	assert.NilError(b, err)

	var mutex sync.Mutex
	leftServer := benchmarkServer{
		PirServer: server,
		b:         b,
		name:      fmt.Sprintf("Left/n=%d,r=%d,CS=%d", config.NumRows, config.RowLen, chunkSize),
		mutex:     &mutex,
	}

	rightServer := benchmarkServer{
		PirServer: server,
		b:         b,
		name:      fmt.Sprintf("Right/n=%d,r=%d,CS=%d", config.NumRows, config.RowLen, chunkSize),
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
	for _, cs := range chunkSizes {
		runPirErasure(b, config.TestConfig, cs)
	}
}

func BenchmarkPirErasureClient(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	db := MakeDB(randSource, config.NumRows, config.RowLen)
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
		fmt.Sprintf("n=%d,r=%d", config.NumRows, config.RowLen),
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pauseServer.b = b
				val, err := client.Read(5)
				assert.NilError(b, err)
				assert.DeepEqual(b, val, db[5])
			}
		})
}

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
