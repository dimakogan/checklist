package boosted

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/rpc"
	"sync"
	"testing"

	"gotest.tools/assert"
)

// For testing server over RPC.
var serverAddr = flag.String("serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")

func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 100)

	leftServer := NewPirServerPunc(RandSource(), db)
	rightServer := NewPirServerPunc(RandSource(), db)
	client := NewPirClientPunc(
		RandSource(),
		len(db),
		[2]PuncPirServer{leftServer, rightServer})
	// Increase number of hints manually to test happy flow
	client.nHints = 100

	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// Test refreshing by reading the same item again
	val, err = client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// Test Batch Read
	readIndices := []int{3, 5, 10}
	vals, errs := client.ReadBatch(readIndices)
	assert.NilError(t, errs[0])
	assert.NilError(t, errs[1])
	assert.NilError(t, errs[2])
	assert.DeepEqual(t, vals[0], db[3])
	assert.DeepEqual(t, vals[1], db[5])
	assert.DeepEqual(t, vals[2], db[10])
}

func TestPIRPuncErasure(t *testing.T) {
	db := MakeDB(256, 100)

	server, err := NewPirServerErasure(RandSource(), db)
	assert.NilError(t, err)
	client, err := NewPirClientErasure(RandSource(), len(db), [2]PuncPirServer{server, server})
	assert.NilError(t, err)
	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRServerOverRPC(t *testing.T) {
	if *serverAddr == "" {
		t.Skip("No remote address flag set. Skipping remote test.")
	}

	// Create a TCP connection to localhost on port 1234
	remote, err := rpc.DialHTTP("tcp", *serverAddr)
	assert.NilError(t, err)

	var none int
	assert.NilError(t, remote.Call("PirRpcServer.SetDBDimensions", DBDimensions{CHUNK_SIZE, 4}, &none))
	assert.NilError(t, remote.Call("PirRpcServer.SetRecordValue", RecordIndexVal{7, Row{'C', 'o', 'o', 'l'}}, &none))

	proxy := NewPirRpcProxy(remote)
	client, err := NewPirClientErasure(RandSource(), CHUNK_SIZE, [2]PuncPirServer{proxy, proxy})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

// Not testing this for now since disabled it
func DontTestPIRPuncKrzysztofTrick(t *testing.T) {
	db := MakeDB(4, 100)
	src := RandSource()

	server := NewPirServerPunc(src, db)

	for i := 0; i < 100; i++ {
		client := NewPirClientPunc(src, len(db), [2]PuncPirServer{server, server})
		// Set nHints to be very high such that the probability of failure due to
		// the index being missing from all of the sets is small
		client.nHints = 100

		assert.NilError(t, client.Init())
		const readIndex = 2
		val, err := client.Read(readIndex)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, db[readIndex])
	}
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
		[]int{
			/*1<<16, 1<<17,1 << 18 , 1<<19, 1 << 20 , 1<<21, 1<<22, 1<<23, 1<<24, 1<<25*/
			//1 << 16, 1 << 17, 1 << 18, 1 << 19, 1 << 20, 1 << 21, 1 << 22, 1 << 23,
			1 << 12, 1 << 14, 1 << 16, 1 << 18, //1 << 20, 1 << 22, 1 << 24,
		}
	dbRecordSize := []int{32}
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

type benchmarkServer struct {
	PuncPirServer
	b    *testing.B
	name string

	// Keep mutex to avoid parallelizm between two "servers" in  tests
	mutex *sync.Mutex
}

func (s *benchmarkServer) Hint(req *HintReq, resp *HintResp) error {
	s.b.Run(
		"Hint/"+s.name,
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := s.PuncPirServer.Hint(req, resp)
				assert.NilError(b, err)
			}
		})
	return s.PuncPirServer.Hint(req, resp)
}

func (s *benchmarkServer) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.b.Run(
		"AnswerBatch/"+s.name,
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := s.PuncPirServer.AnswerBatch(queries, resps)
				assert.NilError(b, err)
			}
		})
	return nil
}

func BenchmarkPirPunc(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)

		server := NewPirServerPunc(randSource, db)
		var mutex sync.Mutex
		benchmarkServer := benchmarkServer{
			PuncPirServer: server,
			b:             b,
			name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:         &mutex,
		}

		client := NewPirClientPunc(randSource, dim.NumRecords, [2]PuncPirServer{&benchmarkServer, &benchmarkServer})
		client.nHints = client.nHints * int(math.Log2(float64(dim.NumRecords)))

		err := client.Init()
		assert.NilError(b, err)

		val, err := client.Read(5)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[5])
	}
}

func BenchmarkPirErasure(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)

		server, err := NewPirServerErasure(randSource, db)
		assert.NilError(b, err)

		var mutex sync.Mutex
		leftServer := benchmarkServer{
			PuncPirServer: server,
			b:             b,
			name:          fmt.Sprintf("Left/n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:         &mutex,
		}

		rightServer := benchmarkServer{
			PuncPirServer: server,
			b:             b,
			name:          fmt.Sprintf("Right/n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:         &mutex,
		}

		client, err := NewPirClientErasure(randSource, dim.NumRecords, [2]PuncPirServer{&leftServer, &rightServer})
		err = client.Init()
		assert.NilError(b, err)

		val, err := client.Read(5)
		assert.NilError(b, err)
		assert.DeepEqual(b, val, db[5])
	}
}

type pauseTimingServer struct {
	PuncPirServer
	b *testing.B

	// Keep mutex to avoid parallelizm between two "servers" in  tests
	mutex *sync.Mutex
}

func (s *pauseTimingServer) Hint(req *HintReq, resp *HintResp) error {
	err := s.PuncPirServer.Hint(req, resp)
	return err
}

func (s *pauseTimingServer) AnswerBatch(queries []QueryReq, resps *[]QueryResp) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.b.StopTimer()
	var err error
	err = s.PuncPirServer.AnswerBatch(queries, resps)
	s.b.StartTimer()
	return err
}

func BenchmarkPirErasureClient(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		server, err := NewPirServerErasure(randSource, db)
		assert.NilError(b, err)

		var mutex sync.Mutex
		pauseServer := pauseTimingServer{
			PuncPirServer: server,
			mutex:         &mutex,
		}

		client, err := NewPirClientErasure(randSource, dim.NumRecords, [2]PuncPirServer{&pauseServer, &pauseServer})
		err = client.Init()
		assert.NilError(b, client.Init())

		b.Run(
			fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
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

func BenchmarkPirErasureRpc(b *testing.B) {
	if *serverAddr == "" {
		b.Skip("No remote address flag set. Skipping remote test.")
	}

	for _, dim := range dbDimensions() {

		// Create a TCP connection to localhost on port 1234
		remote, err := rpc.DialHTTP("tcp", *serverAddr)
		assert.NilError(b, err)

		var none int
		assert.NilError(b, remote.Call("PirRpcServer.SetDBDimensions", dim, &none))

		proxy := NewPirRpcProxy(remote)
		benchmarkServer := benchmarkServer{
			PuncPirServer: proxy,
			b:             b,
			name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
		}

		client, err := NewPirClientErasure(RandSource(), dim.NumRecords, [2]PuncPirServer{&benchmarkServer, proxy})
		assert.NilError(b, err)
		err = client.Init()
		assert.NilError(b, err)

		_, err = client.Read(7)
		assert.NilError(b, err)
	}
}

func BenchmarkHintOnce(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)

	server := NewPirServerPunc(randSource, db)
	benchmarkServer := benchmarkServer{
		PuncPirServer: server,
		b:             b,
		name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
	}

	client := NewPirClientPunc(randSource, dim.NumRecords, [2]PuncPirServer{&benchmarkServer, server})

	err := client.Init()
	assert.NilError(b, err)
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
