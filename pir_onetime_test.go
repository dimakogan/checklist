package boosted

import (
	"fmt"
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func BenchmarkHintOneTime(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, config := range TestConfigs() {
		db := MakeDB(randSource, config.NumRows, config.RowLen)
		client := newPirClientOneTime(randSource, config.NumRows, config.RowLen)
		server := NewPirServerOneTime(randSource, db, config.RowLen)

		hintReq, err := client.RequestHint()
		assert.NilError(b, err)

		b.Run(
			fmt.Sprintf("n=%d,r=%d", config.NumRows, config.RowLen),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var hint pirOneTimeHintResp
					err = server.Hint(*hintReq, &hint)
					assert.NilError(b, err)
				}
			})
	}
}
