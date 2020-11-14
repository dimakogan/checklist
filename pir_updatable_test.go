package boosted

import (
	"testing"

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

	assert.Check(t, len(leftServer.ops) <= len(db)*4)
	assert.Check(t, len(client.ops) <= len(db)*4)
}

func TestPIRUpdatableDefragBetweenUpdates(t *testing.T) {
	db := MakeDB(RandSource(), 20, 100)
	keys := MakeKeys(RandSource(), len(db))

	numDeletesAndAdds := len(db) * 10

	leftServer := NewPirServerUpdatable(RandSource(), Punc)
	rightServer := NewPirServerUpdatable(RandSource(), Punc)

	servers := [2]PirServer{leftServer, rightServer}

	leftServer.AddRows(keys, db)
	rightServer.AddRows(keys, db)

	client := NewPirClientUpdatable(RandSource(), servers)
	assert.NilError(t, client.Init())

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.DeleteRows(keys[0:1])
		rightServer.DeleteRows(keys[0:1])

		assert.NilError(t, client.Update())

		leftServer.AddRows(keys[0:1], db[0:1])
		rightServer.AddRows(keys[0:1], db[0:1])
	}

	assert.Check(t, len(leftServer.ops) <= len(db)*4)
	assert.Check(t, len(client.ops) <= len(db)*4)
}

func TestPIRServerOverRPC(t *testing.T) {
	driver, err := ServerDriver()
	assert.NilError(t, err)

	assert.NilError(t, driver.Configure(TestConfig{
		NumRows:    1000,
		RowLen:     4,
		PresetRows: []RowIndexVal{{7, 0x1234, Row{'C', 'o', 'o', 'l'}}},
		PirType:    Punc,
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
