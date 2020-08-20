package boosted

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"gotest.tools/assert"
)

func BenchmarkHintMatrix(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)

		server := NewPirServerMatrix(randSource, db)
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

		client := NewPIRClient(NewPirClientMatrix(randSource, dim.NumRecords, dim.RecordSize),
			[2]PirServer{&leftServer, &rightServer})

		err := client.Init()
		assert.NilError(b, err)

		readIndex := rand.Intn(len(db))

		val, err := client.Read(readIndex)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[readIndex])
	}
}
