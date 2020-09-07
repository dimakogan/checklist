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

func TestPIRUpdatableHintAndQueryAfterAdditions(t *testing.T) {
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

func TestPIRUpdatableUpdateHintAfterManyChanges(t *testing.T) {
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

func TestPIRUpdatableUpdateHintAfterFewChanges(t *testing.T) {
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
