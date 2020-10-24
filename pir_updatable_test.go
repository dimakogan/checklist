package boosted

import (
	"fmt"
	"math"
	"testing"
	"time"

	"gotest.tools/assert"
)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	testRead(t, keys, db, servers)
}

func TestPIRUpdatableInitAfterFewAdditions(t *testing.T) {
	db := MakeDB(RandSource(), 3000, 100)
	keys := MakeKeys(RandSource(), len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

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
	driver, err := ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    1000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x1234, Row{'C', 'o', 'o', 'l'}}},
		Updatable:  true,
	}, nil))

	//client, err := NewPirClientErasure(RandSource(), 1000, DEFAULT_CHUNK_SIZE, [2]PirServer{proxy, proxy})
	client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

	err = client.Init()
	assert.NilError(t, err)

	val, err := client.Read(0x1234)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, Row("Cool"))
}

func BenchmarkInitial(b *testing.B) {
	driver, err := ServerDriver()
	assert.NilError(b, err)
	rand := RandSource()

	for _, config := range testConfigs() {
		var client PirClient
		var clientInitTime, clientReadTime time.Duration
		var none int
		assert.NilError(b, driver.Configure(config, &none))

		b.Run("Init/"+config.String(), func(b *testing.B) {
			assert.NilError(b, driver.ResetTimers(0, nil))
			for i := 0; i < b.N; i++ {
				if config.Updatable {
					client = NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})
				} else {
					client = NewPIRClient(NewPirClientByType(config.PirType, rand), rand,
						[2]PirServer{driver, driver})
				}
				start := time.Now()
				err = client.Init()
				assert.NilError(b, err)
				clientInitTime += time.Since(start)
			}

			var serverHintTime time.Duration
			assert.NilError(b, driver.GetHintTimer(0, &serverHintTime))
			b.ReportMetric(float64(serverHintTime.Nanoseconds())/float64(b.N), "hint-ns/op")
			b.ReportMetric(float64(clientInitTime.Nanoseconds())/float64(b.N), "init-ns/op")
		})

		b.Run("Read/"+config.String(), func(b *testing.B) {
			assert.NilError(b, driver.ResetTimers(0, nil))
			for i := 0; i < b.N; i++ {
				var rowIV RowIndexVal
				assert.NilError(b, driver.GetRow(rand.Intn(config.NumRows), &rowIV))

				start := time.Now()
				row, err := client.Read(int(rowIV.Key))
				clientReadTime += time.Since(start)
				assert.NilError(b, err)
				assert.DeepEqual(b, row, rowIV.Value)
			}
			var serverAnswerTime time.Duration
			assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))
			b.ReportMetric(float64(serverAnswerTime.Nanoseconds())/float64(b.N), "answer-ns/op")
			b.ReportMetric(float64(clientReadTime.Nanoseconds())/float64(b.N), "read-ns/op")
		})
	}
}

func BenchmarkIncremental(b *testing.B) {
	driver, err := ServerDriver()
	assert.NilError(b, err)
	assert.NilError(b, err)
	rand := RandSource()

	for _, config := range testConfigs() {
		b.Run(config.String(), func(b *testing.B) {
			var none int
			assert.NilError(b, driver.Configure(config, &none))
			client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

			err = client.Init()
			assert.NilError(b, err)

			driver.ResetTimers(0, &none)
			var clientUpdateTime, clientReadTime time.Duration

			for i := 0; i < b.N; i++ {
				changeBatchSize := int(math.Round(math.Sqrt(float64(config.NumRows)))) + 1
				assert.NilError(b, driver.AddRows(changeBatchSize, &none))
				assert.NilError(b, driver.DeleteRows(changeBatchSize, &none))

				start := time.Now()
				client.Update()
				clientUpdateTime += time.Since(start)

				var rowIV RowIndexVal
				assert.NilError(b, driver.GetRow(rand.Intn(config.NumRows), &rowIV))

				start = time.Now()
				row, err := client.Read(int(rowIV.Key))
				clientReadTime += time.Since(start)
				assert.NilError(b, err)
				assert.DeepEqual(b, row, rowIV.Value)

				if *progress {
					fmt.Printf("%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, b.N)
				}
			}
			var serverHintTime, serverAnswerTime time.Duration
			assert.NilError(b, driver.GetHintTimer(0, &serverHintTime))
			assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))

			b.ReportMetric(float64(serverHintTime.Nanoseconds())/float64(b.N), "hint-ns/op")
			b.ReportMetric(float64(serverAnswerTime.Nanoseconds())/float64(b.N), "answer-ns/op")
			b.ReportMetric(float64(clientUpdateTime.Nanoseconds())/float64(b.N), "update-ns/op")
			b.ReportMetric(float64(clientReadTime.Nanoseconds())/float64(b.N), "read-ns/op")
		})
	}
}
