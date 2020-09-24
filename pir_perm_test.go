package boosted

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

func TestPIRPerm(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)

	leftServer := NewPirPermServer(db)
	rightServer := NewPirPermServer(db)
	client := NewPIRClient(
		NewPirPermClient(RandSource()),
		RandSource(),
		[2]PirServer{leftServer, rightServer})

	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func BenchmarkPirPerm(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		dim.NumRecords = 1 << int(math.Ceil(math.Log2(float64(dim.NumRecords))))
		db := MakeDBWithDimensions(randSource, dim)

		server := NewPirPermServer(db)
		var mutex sync.Mutex
		benchmarkServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:     &mutex,
		}

		client := NewPIRClient(
			NewPirPermClient(randSource),
			randSource,
			[2]PirServer{&benchmarkServer, &benchmarkServer})

		err := client.Init()
		assert.NilError(b, err)

		val, err := client.Read(5)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[5])
	}
}
