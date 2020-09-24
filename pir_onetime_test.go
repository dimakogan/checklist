package boosted

import (
	"fmt"
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func BenchmarkHintOneTime(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(randSource, dim)
		client := newPirClientOneTime(randSource, dim.NumRecords, dim.RecordSize)
		server := NewPirServerOneTime(randSource, db, dim.RecordSize)

		hintReq, err := client.RequestHint()
		assert.NilError(b, err)

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var hint pirOneTimeHintResp
					err = server.Hint(*hintReq, &hint)
					assert.NilError(b, err)
				}
			})
	}
}
