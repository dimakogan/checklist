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
	db := MakeDB(256, 100)
	keys := MakeKeys(len(db))

	leftServer := NewPirServerUpdatable(RandSource(), keys, db)
	rightServer := NewPirServerUpdatable(RandSource(), keys, db)
	servers := [2]PirServer{leftServer, rightServer}

	testRead(t, keys, db, servers)
}

func TestPIRUpdatableInitAfterAdditions(t *testing.T) {
	db := MakeDB(3000, 100)
	keys := MakeKeys(len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	rightServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	servers := [2]PirServer{leftServer, rightServer}

	for i := initialSize; i < len(db); i++ {
		leftServer.AddRow(keys[i], db[i])
		rightServer.AddRow(keys[i], db[i])
	}

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
	db := MakeDB(3000, 100)
	keys := MakeKeys(len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	rightServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for i := initialSize; i < len(db); i++ {
		leftServer.AddRow(keys[i], db[i])
		rightServer.AddRow(keys[i], db[i])
	}

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
	db := MakeDB(1200, 100)
	keys := MakeKeys(len(db))

	initialSize := 1000

	leftServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	rightServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for i := initialSize; i < len(db); i++ {
		leftServer.AddRow(keys[i], db[i])
		rightServer.AddRow(keys[i], db[i])
	}

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

	db := MakeDB(finalSize, 100)
	keys := MakeKeys(len(db))

	leftServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	rightServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for s := 0; s < numSteps; s++ {
		for i := initialSize + s*delta; i < initialSize+(s+1)*delta; i++ {
			leftServer.AddRow(keys[i], db[i])
			rightServer.AddRow(keys[i], db[i])
		}

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

	db := MakeDB(initialSize, 100)
	keys := MakeKeys(len(db))

	leftServer := NewPirServerUpdatable(RandSource(), keys, db)
	rightServer := NewPirServerUpdatable(RandSource(), keys, db)
	servers := [2]PirServer{leftServer, rightServer}

	for i := 0; i < deletedPrefix; i++ {
		leftServer.DeleteRow(keys[i])
		rightServer.DeleteRow(keys[i])
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
	db := MakeDB(3000, 100)
	keys := MakeKeys(len(db))

	numDeletes := 1000

	leftServer := NewPirServerUpdatable(RandSource(), keys, db)
	rightServer := NewPirServerUpdatable(RandSource(), keys, db)
	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletes; i++ {
		// interleave deletes
		leftServer.DeleteRow(keys[i*2])
		rightServer.DeleteRow(keys[i*2])
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
	db := MakeDB(20, 100)
	keys := MakeKeys(len(db))

	numDeletesAndAdds := 10
	initialSize := len(db) - numDeletesAndAdds

	leftServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	rightServer := NewPirServerUpdatable(RandSource(), keys[0:initialSize], db[0:initialSize])
	servers := [2]PirServer{leftServer, rightServer}

	client := NewPirClientUpdatable(RandSource(), servers)

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.AddRow(keys[initialSize+i], db[initialSize+i])
		rightServer.AddRow(keys[initialSize+i], db[initialSize+i])

		// interleave deletes from beginning and from new elements
		if i%4 == 0 {
			leftServer.DeleteRow(keys[initialSize+i/2])
			rightServer.DeleteRow(keys[initialSize+i/2])
		} else {
			leftServer.DeleteRow(keys[i*2])
			rightServer.DeleteRow(keys[i*2])
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
	db := MakeDB(2, 100)
	keys := MakeKeys(len(db))

	leftServer := NewPirServerUpdatable(RandSource(), keys, db)
	rightServer := NewPirServerUpdatable(RandSource(), keys, db)
	servers := [2]PirServer{leftServer, rightServer}

	leftServer.DeleteRow(keys[0])
	rightServer.DeleteRow(keys[0])
	leftServer.DeleteRow(keys[1])
	rightServer.DeleteRow(keys[1])

	client := NewPirClientUpdatable(RandSource(), servers)
	assert.NilError(t, client.Init())
}
