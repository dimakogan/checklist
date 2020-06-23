package boosted

import (
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

func TestPIRPunc(t *testing.T) {
	db := MakeDB(256, 10)
	nHints := 256 * int(math.Round(math.Log2(float64(256))))

	server := NewPirServerPunc(RandSource(), db)
	client := NewPirClientPunc(RandSource(), len(db), nHints, server)
	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRPuncErasure(t *testing.T) {
	db := MakeDB(256, 10)

	server := NewPirServerErasure(RandSource(), db)
	client := NewPirClientErasure(RandSource(), len(db), server)
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
	assert.NilError(t, remote.Call("PirRpcServer.SetDBDimensions", DBDimensions{100, 4}, &none))
	assert.NilError(t, remote.Call("PirRpcServer.SetRecordValue", RecordIndexVal{7, Row{'C', 'o', 'o', 'l'}}, &none))

	proxy := NewPirRpcProxy(remote)
	client := NewPirClientPunc(RandSource(), 100, 50, proxy)

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func TestPIRPuncKrzysztofTrick(t *testing.T) {
	db := MakeDB(4, 10)
	// Set nHints to be very high such that the probability of failure due to
	// the index being missing from all of the sets is small
	nHints := 100
	src := RandSource()

	server := NewPirServerPunc(src, db)

	for i := 0; i < 100; i++ {
		client := NewPirClientPunc(src, len(db), nHints, server)
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

type benchmarkServer struct {
	PuncPirServer
	b    *testing.B
	name string
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
	return nil
}

func (s *benchmarkServer) Answer(q QueryReq, resp *QueryResp) error {
	s.b.Run(
		"Answer/"+s.name,
		func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := s.PuncPirServer.Answer(q, resp)
				assert.NilError(b, err)
			}
		})
	return nil
}

func BenchmarkPirPunc(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
		nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))

		server := NewPirServerPunc(randSource, db)
		benchmarkServer := benchmarkServer{
			PuncPirServer: server,
			b:             b,
			name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
		}

		client := NewPirClientPunc(randSource, dim.NumRecords, nHints, &benchmarkServer)

		err := client.Init()
		assert.NilError(b, err)

		_, err = client.Read(5)
		assert.NilError(b, err)
	}
}

func BenchmarkPirPuncRpc(b *testing.B) {
	if *serverAddr == "" {
		b.Skip("No remote address flag set. Skipping remote test.")
	}

	for _, dim := range dbDimensions() {

		// Create a TCP connection to localhost on port 1234
		remote, err := rpc.DialHTTP("tcp", *serverAddr)
		assert.NilError(b, err)

		var none int
		assert.NilError(b, remote.Call("PirRpcServer.SetDBDimensions", dim, &none))

		setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
		nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))

		proxy := NewPirRpcProxy(remote)
		benchmarkServer := benchmarkServer{
			PuncPirServer: proxy,
			b:             b,
			name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
		}

		client := NewPirClientPunc(RandSource(), dim.NumRecords, nHints, &benchmarkServer)

		err = client.Init()
		assert.NilError(b, err)

		_, err = client.Read(7)
		assert.NilError(b, err)
	}
}

func BenchmarkHintErasure(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	for _, dim := range dbDimensions() {
		db := MakeDBWithDimensions(dim)
		server := NewPirServerErasure(randSource, db)
		benchmarkServer := benchmarkServer{
			PuncPirServer: server,
			b:             b,
			name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
		}
		client := NewPirClientErasure(randSource, dim.NumRecords, &benchmarkServer)

		err := client.Init()
		assert.NilError(b, err)
	}
}

func BenchmarkHintOnce(b *testing.B) {
	randSource := rand.New(rand.NewSource(12345))
	dim := DBDimensions{NumRecords: 1024 * 1024, RecordSize: 1024}
	db := MakeDBWithDimensions(dim)
	setSize := int(math.Round(math.Sqrt(float64(dim.NumRecords))))
	nHints := setSize * int(math.Round(math.Log2(float64(dim.NumRecords))))

	server := NewPirServerPunc(randSource, db)
	benchmarkServer := benchmarkServer{
		PuncPirServer: server,
		b:             b,
		name:          fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
	}

	client := NewPirClientPunc(randSource, dim.NumRecords, nHints, &benchmarkServer)

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
