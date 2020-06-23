package boosted

import (
	"bytes"
	"flag"
	"fmt"
	"math"
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
	assert.DeepEqual(t, val, db[readIndex])
}

// TestPIRStub is a simple end-to-end test.
func TestPIRStub(t *testing.T) {
	db := MakeDB(256, 1024)
	client := NewPirClientStub()
	server := PIRServerStub{db: db}

	testBasicRead(t, db, client, server)
}

func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 10)
	nHints := 256 * int(math.Round(math.Log2(float64(256))))

	client := newPirClientPunc(RandSource(), len(db), nHints)

	server := NewPirServerPunc(RandSource(), db)
	t.Run(
		"Hint",
		func(t *testing.T) {
			testBasicRead(t, db, client, server)
		})
}

func TestPIRServerOverRPC(t *testing.T) {
	if *serverAddr == "" {
		t.Skip("No remote address flag set. Skipping remote test.")
	}

	// Create a TCP connection to localhost on port 1234
	remote, err := rpc.DialHTTP("tcp", *serverAddr)
	assert.NilError(t, err)

	var none int
	assert.NilError(t, remote.Call("PIRServer.SetDBDimensions", DBDimensions{100, 4}, &none))
	assert.NilError(t, remote.Call("PIRServer.SetRecordValue", RecordIndexVal{7, Row{'C', 'o', 'o', 'l'}}, &none))

	pir := newPirClientPunc(RandSource(), 100, 10)
	assert.Assert(t, pir != nil)
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
	numDBRecords :=
		//[]int{2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576}
		[]int{ /*1<<16, 1<<17,*/ 1 << 18 /* 1<<19, 1<<20, 1<<21, 1<<22, 1<<23, 1<<24, 1<<25*/}
	dbRecordSize := []int{96}
	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(2 * 1024 * 1024 * 1024)

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
		setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
		nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))

		client := newPirClientPunc(randSource, dim.NumRecords, nHints)
		server := NewPirServerPunc(randSource, db)

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

func BenchmarkHintErasure(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		client := newPirClientErasure(randSource, dim.NumRecords)
		server := NewPirServerErasure(randSource, db)

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

func BenchmarkHintOnce(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)
	setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
	nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))
	client := newPirClientPunc(randSource, dim.NumRecords, nHints)

	hintReq, err := client.RequestHint()
	assert.NilError(b, err)

	server := NewPirServerPunc(randSource, db)
	b.Run(
		"hint",
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var hint HintResp
				err = server.Hint(hintReq, &hint)
				assert.NilError(b, err)
			}
		})
}

func BenchmarkHintMatrix(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		client := newPirClientMatrix(randSource, dim.NumRecords, dim.RecordSize)
		server := NewPirServerMatrix(randSource, db, dim.RecordSize)

		hintReq, err := client.RequestHint()
		assert.NilError(b, err)

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var hint pirMatrixHintResp
					err = server.Hint(*hintReq, &hint)
					assert.NilError(b, err)
				}
			})
	}
}

func BenchmarkHintOneTime(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
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

func BenchmarkNothingRandom(b *testing.B) {
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	nHints := 1024
	setLen := 1024

	out := make(Row, dim.RecordSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				q := ((123124124 * k) + 912812367) % dim.NumRecords
				xorInto(out, db[q])
			}
		}
	}
}

func BenchmarkNothingLinear(b *testing.B) {
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	nHints := 1024
	setLen := 1024

	out := make(Row, dim.RecordSize)
	b.ResetTimer()
	q := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < nHints; j++ {
			for k := 0; k < setLen; k++ {
				xorInto(out, db[q])
				q = (q + 1) % dim.NumRecords
			}
		}
	}
}

func BenchmarkAnswer(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
		nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))
		client := newPirClientPunc(randSource, dim.NumRecords, nHints)
		server := NewPirServerPunc(randSource, db)

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

func BenchmarkAnswerOverRPC(b *testing.B) {
	if *serverAddr == "" {
		b.Skip("No remote address flag set. Skipping remote test.")
	}

	randSource := RandSource()
	for _, dim := range dbDimensions() {

		// Create a TCP connection to localhost on port 1234
		remote, err := rpc.DialHTTP("tcp", *serverAddr)
		assert.NilError(b, err)

		var none int
		assert.NilError(b, remote.Call("PIRServer.SetDBDimensions", dim, &none))

		// Prepare a bunch of queries to avoid hitting cache effects on the server.

		preparedValues := make([]Row, 10)
		for i := 0; i < len(preparedValues); i++ {
			preparedValues[i] = make([]byte, dim.RecordSize)
			randSource.Read(preparedValues[i])
			assert.NilError(b, remote.Call("PIRServer.SetRecordValue", RecordIndexVal{i, preparedValues[i]}, &none))
		}

		// Initialize clients with valid hints
		preparedClients := make([]PIRClient, len(preparedValues))
		preparedQueries := make([]*QueryReq, len(preparedValues))
		for i := 0; i < len(preparedClients); i++ {
			setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
			nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))
			preparedClients[i] = newPirClientPunc(randSource, dim.NumRecords, nHints)
			hintReq, err := preparedClients[i].RequestHint()
			assert.NilError(b, err)

			var hintResp HintResp
			assert.NilError(b, remote.Call("PIRServer.Hint", hintReq, &hintResp))
			assert.Assert(b, len(hintResp.Hints) > 0)
			assert.NilError(b, preparedClients[i].InitHint(&hintResp))

			queries, err := preparedClients[i].Query(i)
			assert.NilError(b, err)
			preparedQueries[i] = queries[0]
		}
		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var queryResp QueryResp
					assert.NilError(b, remote.Call("PIRServer.Answer", preparedQueries[i%len(preparedQueries)], &queryResp))
					val, err := preparedClients[i%len(preparedQueries)].Reconstruct([]*QueryResp{&queryResp})
					assert.NilError(b, err)
					assert.Equal(b, val, preparedValues[i%len(preparedValues)])
				}
			})
	}
}
