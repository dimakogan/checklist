package boosted

import (
  "bytes"
	"fmt"
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func makeDb(nRows int, rowLen int) []Row {
  db := make([]Row, nRows)
  src := randSource()
  for i := range db {
    db[i] = make([]byte, rowLen)
    src.Read(db[i])
  }
  return db
}

func testBasicRead(t *testing.T, db []Row, client PIRClient, server PIRServer) {
	hintReq, err := client.RequestHint()
	assert.NilError(t, err)
	var hintResp HintResp
	err = server.Hint(hintReq, &hintResp)
	assert.NilError(t, err)
	assert.NilError(t, client.InitHint(&hintResp))

	const readIndex = 2
	queryReq, err := client.Query(readIndex)
	assert.NilError(t, err)

	var queryResp QueryResp
	err = server.Answer(queryReq[0], &queryResp)
	assert.NilError(t, err)
	val, err := client.Reconstruct([]*QueryResp{&queryResp})
	assert.NilError(t, err)
  assert.Assert(t, bytes.Equal(val, db[readIndex]))
}

// TestPIRStub is a simple end-to-end test.
func TestPIRStub(t *testing.T) {
  db := makeDb(100, 1024)
	client := NewPirClientStub()
	server := PIRServerStub{db: db}

	testBasicRead(t, db, client, server)
}


func TestPIRPunc(t *testing.T) {
  db := makeDb(100, 1024)
	client := newPirClientPunc(randSource(), len(db))
	server := newPirServerPunc(randSource(), db)

	testBasicRead(t, db, client, server)
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

func BenchmarkServer(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	numDBRecords := []int{1000, 10 * 1000, 100 * 1000, 1000 * 1000}
	dbRecordSize := []int{10, 100, 1000, 10 * 1000, 100 * 1000}
	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(2 * 1000 * 1000 * 1000)

	for _, n := range numDBRecords {
		for _, recSize := range dbRecordSize {
			if int64(n)*int64(recSize) > maxDBSizeBytes {
				continue
			}
      db := makeDb(n, recSize)
			client := newPirClientPunc(randSource, n)
			server := newPirServerPunc(randSource, db)

			hintReq, err := client.RequestHint()
			assert.NilError(b, err)

			b.Run(fmt.Sprintf("HintGeneration/n=%d,B=%d/", n, recSize), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var hint HintResp
					err = server.Hint(hintReq, &hint)
					assert.NilError(b, err)
				}
			})

			queryReq, err := client.Query(5)
			assert.NilError(b, err)

			b.Run(fmt.Sprintf("QueryAnswer/n=%d,B=%d/", n, recSize), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var resp QueryResp
					err = server.Answer(queryReq[0], &resp)
					assert.NilError(b, err)
				}
			})

		}
	}

}
