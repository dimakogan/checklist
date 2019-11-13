package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

func testBasicRead(t *testing.T, db []string, client PIRClient, server PIRServer) {
	hintReq, err := client.RequestHint()
	assert.NilError(t, err)
	hintResp, err := server.Hint(hintReq)
	assert.NilError(t, err)
	assert.NilError(t, client.InitHint(hintResp))

	const readIndex = 2
	queryReq, err := client.Query(readIndex)
	assert.NilError(t, err)

	queryResp, err := server.Answer(queryReq[0])
	assert.NilError(t, err)
	val, err := client.Reconstruct([]*QueryResp{queryResp})
	assert.NilError(t, err)
	assert.Equal(t, val, db[readIndex])
}

// TestPIRStub is a simple end-to-end test.
func TestPIRStub(t *testing.T) {
	var db = []string{"A", "B", "C", "D"}

	client := newPirClientStub()
	server := pirServerStub{db: db}

	testBasicRead(t, db, client, server)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Save results to avoid compiler optimizations.
var hint *HintResp
var resp *QueryResp

func BenchmarkServer(b *testing.B) {
	numDBRecords := []int{1000, 10 * 1000, 100 * 1000, 1000 * 1000}
	dbRecordSize := []int{10, 100, 1000, 10 * 1000, 100 * 1000}
	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(2 * 1000 * 1000 * 1000)

	for _, n := range numDBRecords {
		for _, recSize := range dbRecordSize {
			if int64(n)*int64(recSize) > maxDBSizeBytes {
				continue
			}
			var db = make([]string, n)
			for i := 0; i < n; i++ {
				db[i] = randStringBytes(recSize)
			}
			client := newPirClientStub()
			server := pirServerStub{db: db, numReadsOnHint: n, numReadsOnAnswer: int(math.Floor(math.Sqrt(float64(n))))}

			hintReq, err := client.RequestHint()
			assert.NilError(b, err)

			b.Run(fmt.Sprintf("HintGeneration/n=%d,B=%d", n, recSize), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					hint, err = server.Hint(hintReq)
					assert.NilError(b, err)
				}
			})

			queryReq, err := client.Query(5)
			assert.NilError(b, err)

			b.Run(fmt.Sprintf("QueryAnswer/n=%d,B=%d", n, recSize), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					resp, err = server.Answer(queryReq[0])
					assert.NilError(b, err)
				}
			})

		}
	}

}
