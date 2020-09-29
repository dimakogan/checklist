package boosted

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

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
	for _, config := range testConfigs() {
		for _, cs := range chunkSizes {
			runPirErasure(b, config, cs)
		}
	}
}

func BenchmarkPirErasureClient(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, config := range testConfigs() {
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

}
