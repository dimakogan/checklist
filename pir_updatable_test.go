package boosted

import (
	"flag"
	"fmt"
	"math"
	"net/rpc"
	"strings"
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"
)

// For testing server over RPC.
var serverAddr = flag.String("serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")
var updatablePirType = flag.String("pirType", PirPuncturable.String(), fmt.Sprintf("Updatable PIR type: [%s]", strings.Join(PirTypeStrings(), "|")))

func updatableServer() (PirServerDriver, error) {
	if *serverAddr != "" {
		// Create a TCP connection to localhost on port 1234
		remote, err := rpc.DialHTTP("tcp", *serverAddr)
		if err != nil {
			return nil, err
		}
		return NewPirRpcProxy(remote), nil
	} else {
		return NewPirServerDriver()
	}
}

func testRead(t *testing.T, keys []uint32, db []Row, servers [2]PirServer) {
	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// Test refreshing by reading the same item again
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableStatic(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)
	keys := MakeKeys(RandSource(), len(db))

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	testRead(t, keys, db, servers)
}

func TestPIRUpdatableInitAfterFewAdditions(t *testing.T) {
	db := MakeDB(RandSource(), 3000, 100)
	keys := MakeKeys(RandSource(), len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}
	leftServer.AddRows(keys[0:initialSize], db[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], db[0:initialSize])
	leftServer.AddRows(keys[initialSize:], db[initialSize:])
	rightServer.AddRows(keys[initialSize:], db[initialSize:])

	client := NewPirClientUpdatable(RandSource(), servers)

	// Read something from the beginning
	assert.NilError(t, client.Init())
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	readIndex = len(db) - 100
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableUpdateAfterManyAdditions(t *testing.T) {
	db := MakeDB(RandSource(), 3000, 100)
	keys := MakeKeys(RandSource(), len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	leftServer.AddRows(keys[0:initialSize], db[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], db[0:initialSize])

	assert.NilError(t, client.Init())

	leftServer.AddRows(keys[initialSize:], db[initialSize:])
	rightServer.AddRows(keys[initialSize:], db[initialSize:])

	assert.NilError(t, client.Update())
	// Read something from the beginning
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	readIndex = len(db) - 100
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableUpdateAfterFewAdditions(t *testing.T) {
	db := MakeDB(RandSource(), 1200, 100)
	keys := MakeKeys(RandSource(), len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	leftServer.AddRows(keys[0:initialSize], db[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], db[0:initialSize])

	assert.NilError(t, client.Init())

	leftServer.AddRows(keys[initialSize:], db[initialSize:])
	rightServer.AddRows(keys[initialSize:], db[initialSize:])

	assert.NilError(t, client.Update())
	// Read something from the beginning
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	readIndex = len(db) - 100
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableMultipleUpdates(t *testing.T) {
	initialSize := 1000
	delta := 200
	numSteps := 10

	finalSize := initialSize + delta*numSteps

	db := MakeDB(RandSource(), finalSize, 100)
	keys := MakeKeys(RandSource(), len(db))

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	leftServer.AddRows(keys[0:initialSize], db[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], db[0:initialSize])

	assert.NilError(t, client.Init())

	for s := 0; s < numSteps; s++ {
		start := initialSize + s*delta
		end := initialSize + (s+1)*delta
		leftServer.AddRows(keys[start:end], db[start:end])
		rightServer.AddRows(keys[start:end], db[start:end])

		assert.NilError(t, client.Update())
		// Read something from the beginning
		readIndex := 2
		key := keys[readIndex]
		val, err := client.Read(int(key))
		assert.NilError(t, err)
		assert.DeepEqual(t, val, db[readIndex])

		readIndex = initialSize + (s+1)*delta - 100
		key = keys[readIndex]
		val, err = client.Read(int(key))
		assert.NilError(t, err)
		assert.DeepEqual(t, val, db[readIndex])
	}
}

func TestPIRUpdatableInitAfterDeletes(t *testing.T) {
	initialSize := 500
	deletedPrefix := 200

	db := MakeDB(RandSource(), initialSize, 100)
	keys := MakeKeys(RandSource(), len(db))

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	for i := 0; i < deletedPrefix; i++ {
		leftServer.DeleteRows(keys[i : i+1])
		rightServer.DeleteRows(keys[i : i+1])
	}

	client := NewPirClientUpdatable(RandSource(), servers)
	assert.NilError(t, client.Init())

	// Check that reading a deleted element fails
	readIndex := deletedPrefix - 1
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.ErrorContains(t, err, "")

	readIndex = deletedPrefix + 100
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableUpdateAfterDeletes(t *testing.T) {
	db := MakeDB(RandSource(), 3000, 100)
	keys := MakeKeys(RandSource(), len(db))

	numDeletes := 1000

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletes; i++ {
		// interleave deletes
		leftServer.DeleteRows([]uint32{keys[i*2]})
		rightServer.DeleteRows([]uint32{keys[i*2]})
	}

	assert.NilError(t, client.Update())

	readIndex := 100
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.ErrorContains(t, err, "")

	readIndex = 101
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}

func TestPIRUpdatableUpdateAfterAddsAndDeletes(t *testing.T) {
	db := MakeDB(RandSource(), 20, 100)
	keys := MakeKeys(RandSource(), len(db))

	numDeletesAndAdds := 10
	initialSize := len(db) - numDeletesAndAdds

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	leftServer.AddRows(keys[0:initialSize], db[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], db[0:initialSize])

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.AddRows([]uint32{keys[initialSize+i]}, []Row{db[initialSize+i]})
		rightServer.AddRows([]uint32{keys[initialSize+i]}, []Row{db[initialSize+i]})

		// interleave deletes from beginning and from new elements
		if i%4 == 0 {
			leftServer.DeleteRows([]uint32{keys[initialSize+i/2]})
			rightServer.DeleteRows([]uint32{keys[initialSize+i/2]})
		} else {
			leftServer.DeleteRows([]uint32{keys[i*2]})
			rightServer.DeleteRows([]uint32{keys[i*2]})
		}
	}

	assert.NilError(t, client.Update())

	// deleted element from the original elements
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(int(key))
	assert.ErrorContains(t, err, "")

	// deleted element from the added elements
	readIndex = initialSize + 2
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.ErrorContains(t, err, "")

	// original non-deleted element
	readIndex = 3
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

	// added non-deleted element
	readIndex = initialSize + 3
	key = keys[readIndex]
	val, err = client.Read(int(key))
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])

}

func TestPIRUpdatableDeleteAll(t *testing.T) {
	db := MakeDB(RandSource(), 2, 100)
	keys := MakeKeys(RandSource(), len(db))

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.DeleteRows([]uint32{keys[0]})
	rightServer.DeleteRows([]uint32{keys[0]})
	leftServer.DeleteRows([]uint32{keys[1]})
	rightServer.DeleteRows([]uint32{keys[1]})

	client := NewPirClientUpdatable(RandSource(), servers)
	assert.NilError(t, client.Init())
}

func TestPIRUpdatableDefrag(t *testing.T) {
	db := MakeDB(RandSource(), 20, 100)
	keys := MakeKeys(RandSource(), len(db))

	numDeletesAndAdds := len(db) * 10

	leftServer := NewPirServerUpdatable(RandSource(), PirPuncturable)
	rightServer := NewPirServerUpdatable(RandSource(), PirPuncturable)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.DeleteRows(keys[0:1])
		rightServer.DeleteRows(keys[0:1])

		leftServer.AddRows(keys[0:1], db[0:1])
		rightServer.AddRows(keys[0:1], db[0:1])
	}

	client := NewPirClientUpdatable(RandSource(), servers)
	assert.NilError(t, client.Init())

	assert.Check(t, len(leftServer.timedRows) <= len(db)*4)
}

func TestPIRServerOverRPC(t *testing.T) {
	driver, err := updatableServer()
	assert.NilError(t, err)

	var none int
	assert.NilError(t, driver.ResetDBDimensions(DBDimensions{1000, 4}, &none))
	assert.NilError(t, driver.SetRecordValue(RecordIndexVal{7, 0x1234, Row{'C', 'o', 'o', 'l'}}, &none))

	//client, err := NewPirClientErasure(RandSource(), 1000, DEFAULT_CHUNK_SIZE, [2]PirServer{proxy, proxy})
	client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x1234)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func BenchmarkPirUpdatableInitial(b *testing.B) {
	driver, err := updatableServer()
	assert.NilError(b, err)
	pirType, err := PirTypeString(*updatablePirType)
	assert.NilError(b, err)

	for _, dim := range dbDimensions() {

		var none int
		assert.NilError(b, driver.SetPIRType(pirType, &none))
		assert.NilError(b, driver.ResetDBDimensions(dim, &none))
		assert.NilError(b, driver.SetRecordValue(RecordIndexVal{7, 0x1234, make([]byte, dim.RecordSize)}, &none))

		var mutex sync.Mutex
		benchmarkServer := benchmarkServer{
			PirServer: driver,
			b:         b,
			name:      fmt.Sprintf("n=%d,B=%d", dim.NumRecords, dim.RecordSize),
			mutex:     &mutex,
		}

		client := NewPirClientUpdatable(RandSource(), [2]PirServer{&benchmarkServer, driver})
		err = client.Init()
		assert.NilError(b, err)

		var record RecordIndexVal
		driver.GetRecord(7, &record)
		row, err := client.Read(int(record.Key))
		assert.NilError(b, err)
		assert.DeepEqual(b, row, record.Value)
	}
}

func BenchmarkUpdatablePirHint(b *testing.B) {
	driver, err := updatableServer()
	assert.NilError(b, err)
	pirType, err := PirTypeString(*updatablePirType)
	assert.NilError(b, err)

	for _, dim := range dbDimensions() {
		var none int
		assert.NilError(b, driver.SetPIRType(pirType, &none))
		assert.NilError(b, driver.ResetDBDimensions(dim, &none))
		client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

		err = client.Init()
		assert.NilError(b, err)

		var clientAndServerHintTime, clientAndServerAnswerTime, serverHintTime time.Duration

		changeBatchSize := int(math.Sqrt(float64(dim.NumRecords))) / 2
		//numChanges := 3 * dim.NumRecords
		numBatches := 100 //numChanges / changeBatchSize
		for i := 0; i < numBatches; i++ {
			assert.NilError(b, driver.AddRows(changeBatchSize, &none))
			assert.NilError(b, driver.DeleteRows(changeBatchSize, &none))

			startTime := time.Now()
			client.Update()
			clientAndServerHintTime += time.Since(startTime)
		}

		var record RecordIndexVal
		assert.NilError(b, driver.GetRecord(7, &record))
		startTime := time.Now()
		row, err := client.Read(int(record.Key))
		clientAndServerAnswerTime += time.Since(startTime)

		assert.NilError(b, err)
		assert.DeepEqual(b, row, record.Value)

		assert.NilError(b, driver.GetHintTimer(0, &serverHintTime))

		b.ReportMetric(float64(clientAndServerHintTime.Nanoseconds())/float64(numBatches), "total-ns/hint")
		b.ReportMetric(float64(serverHintTime.Nanoseconds())/float64(numBatches), "server-ns/hint")

		// var serverAnswerTime int
		//assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))
		// b.ReportMetric(float64(clientAndServerAnswerTime.Nanoseconds()), "ns/query")
		// b.ReportMetric(float64(serverAnswerTime.Nanoseconds()/2), "server-ns/query")
	}
}

func BenchmarkUpdatablePir(b *testing.B) {
	driver, err := updatableServer()
	assert.NilError(b, err)
	pirType, err := PirTypeString(*updatablePirType)

	for _, dim := range dbDimensions() {

		var none int
		assert.NilError(b, driver.SetPIRType(pirType, &none))
		assert.NilError(b, driver.ResetDBDimensions(dim, &none))
		client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

		err = client.Init()
		assert.NilError(b, err)

		var serverAnswerTime time.Duration

		changeBatchSize := int(math.Sqrt(float64(dim.NumRecords)) * 9 / 10)
		numBatches := 10
		for i := 0; i < numBatches; i++ {
			assert.NilError(b, driver.AddRows(changeBatchSize, &none))
			assert.NilError(b, driver.DeleteRows(changeBatchSize, &none))
		}

		assert.NilError(b, client.Update())

		var record RecordIndexVal
		assert.NilError(b, driver.GetRecord(7, &record))

		b.Run("Answer", func(b *testing.B) {
			b.StopTimer()
			assert.NilError(b, driver.ResetTimers(0, nil))
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				row, err := client.Read(int(record.Key))
				assert.NilError(b, err)
				assert.DeepEqual(b, row, record.Value)
			}
			b.StopTimer()
			assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))
			// Divide by 2 to get per-server time
			b.ReportMetric(float64(serverAnswerTime.Nanoseconds())/float64(b.N)/2, "server-ns/query")
			b.StartTimer()
		})
	}
}
