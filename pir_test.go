package boosted

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"net/rpc"
	"testing"

	"gotest.tools/assert"
)

// For testing server over RPC.
var serverAddr = flag.String("serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")

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
	db := MakeDB(100, 1024)
	client := NewPirClientStub()
	server := PIRServerStub{db: db}

	testBasicRead(t, db, client, server)
}

func TestPIRPunc(t *testing.T) {
	db := MakeDB(100, 1024)
	client := newPirClientPunc(RandSource(), len(db))
	server := newPirServerPunc(RandSource(), db)

	testBasicRead(t, db, client, server)
}

func TestPIRServerOverRPC(t *testing.T) {
	if *serverAddr == "" {
		t.Skip("No remote address flag set. Skipping remote test.")
	}

	pir := NewPirClientStub()
	assert.Assert(t, pir != nil)

	// Create a TCP connection to localhost on port 1234
	remote, err := rpc.DialHTTP("tcp", *serverAddr)
	assert.NilError(t, err)

	var none int
	assert.NilError(t, remote.Call("PIRServer.SetDBDimensions", DBDimensions{100, 4}, &none))
	assert.NilError(t, remote.Call("PIRServer.SetRecordValue", RecordIndexVal{7, Row{'C', 'o', 'o', 'l'}}, &none))

	client, err := NewRpcPirClient(remote, pir)
	assert.NilError(t, err)

	val, err := client.Read(7)
	assert.NilError(t, err)
	assert.Assert(t, bytes.Equal(val, []byte("Cool")))
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

func dbDimensions() []DBDimensions {
	var dims []DBDimensions
	numDBRecords := []int{1000, 10 * 1000, 100 * 1000, 1000 * 1000}
	dbRecordSize := []int{10, 100, 1000, 10 * 1000, 100 * 1000}
	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(2 * 1000 * 1000 * 1000)

	for _, n := range numDBRecords {
		for _, recSize := range dbRecordSize {
			if int64(n)*int64(recSize) > maxDBSizeBytes {
				continue
			}
			dims = append(dims, DBDimensions{NumRecords: n, RecordSize: recSize})
		}
	}
	return dims
}

func BenchmarkHint(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		client := newPirClientPunc(randSource, dim.NumRecords)
		server := newPirServerPunc(randSource, db)

		hintReq, err := client.RequestHint()
		assert.NilError(b, err)

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var hint HintResp
					err = server.Hint(hintReq, &hint)
					assert.NilError(b, err)
				}
			})
	}
}

func BenchmarkAnswer(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		client := newPirClientPunc(randSource, dim.NumRecords)
		server := newPirServerPunc(randSource, db)

		// Initialize client with valid hint
		hintReq, err := client.RequestHint()
		assert.NilError(b, err)
		var hint HintResp
		err = server.Hint(hintReq, &hint)
		assert.NilError(b, err)
		assert.NilError(b, client.InitHint(&hint))

		// Prepare a bunch of queries to avoid hitting cache effects on the server.
		preparedQueries := make([]*QueryReq, 10)
		for i := 0; i < len(preparedQueries); i++ {
			queries, err := client.Query(5)
			assert.NilError(b, err)
			preparedQueries[i] = queries[0]
		}

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var resp QueryResp
					err = server.Answer(preparedQueries[i%len(preparedQueries)], &resp)
					assert.NilError(b, err)
				}
			})
	}
}
