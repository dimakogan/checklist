package updatable

import (
	"testing"

	"checklist/pir"

	"gotest.tools/assert"
)

func testRead(t *testing.T, keys []uint32, rows []pir.Row, servers [2]UpdatableServer) {
	client := NewClient(pir.RandSource(), pir.Punc, servers)

	assert.NilError(t, client.Init())
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

	// Test refreshing by reading the same item again
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableStatic(t *testing.T) {
	keys, rows := pir.MakeKeysRows(256, 100)

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)

	testRead(t, keys, rows, servers)
}

func TestPIRUpdatableInitAfterFewAdditions(t *testing.T) {
	keys, rows := pir.MakeKeysRows(3000, 100)

	initialSize := 1000

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}
	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	leftServer.AddRows(keys[initialSize:], rows[initialSize:])
	rightServer.AddRows(keys[initialSize:], rows[initialSize:])

	client := NewClient(pir.RandSource(), pir.Punc, servers)

	// Read something from the beginning
	assert.NilError(t, client.Init())
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

	readIndex = len(rows) - 100
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableUpdateAfterManyAdditions(t *testing.T) {
	keys, rows := pir.MakeKeysRows(3000, 100)

	initialSize := 1000

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	client := NewClient(pir.RandSource(), pir.Punc, servers)

	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])

	assert.NilError(t, client.Init())

	leftServer.AddRows(keys[initialSize:], rows[initialSize:])
	rightServer.AddRows(keys[initialSize:], rows[initialSize:])

	assert.NilError(t, client.Update())
	// Read something from the beginning
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

	readIndex = len(rows) - 100
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableUpdateAfterFewAdditions(t *testing.T) {
	keys, rows := pir.MakeKeysRows(1200, 100)

	initialSize := 1000

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10

	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])

	assert.NilError(t, client.Init())

	leftServer.AddRows(keys[initialSize:], rows[initialSize:])
	rightServer.AddRows(keys[initialSize:], rows[initialSize:])

	assert.NilError(t, client.Update())
	// Read something from the beginning
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

	readIndex = len(rows) - 100
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableMultipleUpdates(t *testing.T) {
	initialSize := 1000
	delta := 200
	numSteps := 10

	finalSize := initialSize + delta*numSteps
	keys, rows := pir.MakeKeysRows(finalSize, 100)

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10

	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])

	assert.NilError(t, client.Init())

	for s := 0; s < numSteps; s++ {
		start := initialSize + s*delta
		end := initialSize + (s+1)*delta
		leftServer.AddRows(keys[start:end], rows[start:end])
		rightServer.AddRows(keys[start:end], rows[start:end])

		assert.NilError(t, client.Update())
		// Read something from the beginning
		readIndex := 2
		key := keys[readIndex]
		val, err := client.Read(key)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, rows[readIndex])

		readIndex = initialSize + (s+1)*delta - 100
		key = keys[readIndex]
		val, err = client.Read(key)
		assert.NilError(t, err)
		assert.DeepEqual(t, val, rows[readIndex])
	}
}

func TestPIRUpdatableInitAfterDeletes(t *testing.T) {
	initialSize := 500
	deletedPrefix := 200

	keys, rows := pir.MakeKeysRows(initialSize, 100)

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)

	for i := 0; i < deletedPrefix; i++ {
		leftServer.DeleteRows(keys[i : i+1])
		rightServer.DeleteRows(keys[i : i+1])
	}

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10
	assert.NilError(t, client.Init())

	// Check that reading a deleted element fails
	readIndex := deletedPrefix - 1
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.ErrorContains(t, err, "")

	readIndex = deletedPrefix + 100
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableUpdateAfterDeletes(t *testing.T) {
	keys, rows := pir.MakeKeysRows(3000, 100)

	numDeletes := 1000

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletes; i++ {
		// interleave deletes
		leftServer.DeleteRows([]uint32{keys[i*2]})
		rightServer.DeleteRows([]uint32{keys[i*2]})
	}

	assert.NilError(t, client.Update())

	readIndex := 100
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.ErrorContains(t, err, "")

	readIndex = 101
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])
}

func TestPIRUpdatableUpdateAfterAddsAndDeletes(t *testing.T) {
	keys, rows := pir.MakeKeysRows(20, 100)

	numDeletesAndAdds := 10
	initialSize := len(rows) - numDeletesAndAdds

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10

	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])

	assert.NilError(t, client.Init())

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.AddRows([]uint32{keys[initialSize+i]}, []pir.Row{rows[initialSize+i]})
		rightServer.AddRows([]uint32{keys[initialSize+i]}, []pir.Row{rows[initialSize+i]})

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
	val, err := client.Read(key)
	assert.ErrorContains(t, err, "")

	// deleted element from the added elements
	readIndex = initialSize + 2
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.ErrorContains(t, err, "")

	// original non-deleted element
	readIndex = 3
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

	// added non-deleted element
	readIndex = initialSize + 3
	key = keys[readIndex]
	val, err = client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[readIndex])

}

func TestPIRUpdatableDeleteAll(t *testing.T) {
	keys, rows := pir.MakeKeysRows(2, 100)

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)
	leftServer.DeleteRows([]uint32{keys[0]})
	rightServer.DeleteRows([]uint32{keys[0]})
	leftServer.DeleteRows([]uint32{keys[1]})
	rightServer.DeleteRows([]uint32{keys[1]})

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10
	assert.NilError(t, client.Init())
}

func TestPIRUpdatableDefrag(t *testing.T) {
	keys, rows := pir.MakeKeysRows(20, 100)

	numDeletesAndAdds := len(rows) * 10

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.DeleteRows(keys[0:1])
		rightServer.DeleteRows(keys[0:1])

		leftServer.AddRows(keys[0:1], rows[0:1])
		rightServer.AddRows(keys[0:1], rows[0:1])
	}

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10
	assert.NilError(t, client.Init())

	assert.Check(t, len(leftServer.ops) <= len(rows)*4)
	assert.Check(t, len(client.ops) <= len(rows)*4)
}

func TestPIRUpdatableDefragBetweenUpdates(t *testing.T) {
	keys, rows := pir.MakeKeysRows(20, 100)

	numDeletesAndAdds := len(rows) * 10

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}

	leftServer.AddRows(keys, rows)
	rightServer.AddRows(keys, rows)

	client := NewClient(pir.RandSource(), pir.Punc, servers)
	client.waterfall.smallestLayerSizeOverride = 10
	assert.NilError(t, client.Init())

	for i := 0; i < numDeletesAndAdds; i++ {
		leftServer.DeleteRows(keys[0:1])
		rightServer.DeleteRows(keys[0:1])

		assert.NilError(t, client.Update())

		leftServer.AddRows(keys[0:1], rows[0:1])
		rightServer.AddRows(keys[0:1], rows[0:1])
	}

	assert.Check(t, len(leftServer.ops) <= len(rows)*4)
	assert.Check(t, len(client.ops) <= len(rows)*4)
}

func TestPIRUpdatableInitAfterKeyOverride(t *testing.T) {
	keys, rows := pir.MakeKeysRows(3000, 100)

	initialSize := 1000

	leftServer := NewUpdatableServer()
	rightServer := NewUpdatableServer()

	servers := [2]UpdatableServer{leftServer, rightServer}
	leftServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[0:initialSize])
	leftServer.AddRows(keys[0:initialSize], rows[initialSize:2*initialSize])
	rightServer.AddRows(keys[0:initialSize], rows[initialSize:2*initialSize])

	client := NewClient(pir.RandSource(), pir.Punc, servers)

	// Read something from the beginning
	assert.NilError(t, client.Init())
	readIndex := 2
	key := keys[readIndex]
	val, err := client.Read(key)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, rows[initialSize+readIndex])
}
