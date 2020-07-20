package boosted

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

func TestPIRPerm(t *testing.T) {
	db := MakeDB(256, 100)

	leftServer := NewPirPermServer(db)
	rightServer := NewPirPermServer(db)
	client := NewPirPermClient(
		RandSource(),
		len(db),
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
		db := MakeDBWithDimensions(dim)

		server := NewPirPermServer(db)
		var mutex sync.Mutex
		benchmarkServer := benchmarkServer{
			PirServer: server,
			b:         b,
			name:      fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:     &mutex,
		}

		client := NewPirPermClient(randSource, dim.NumRecords, [2]PirServer{&benchmarkServer, &benchmarkServer})

		err := client.Init()
		assert.NilError(b, err)

		val, err := client.Read(5)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[5])
	}
}
