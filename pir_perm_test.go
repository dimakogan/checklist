package boosted

import (
	"testing"

	"gotest.tools/assert"
)

func TestPIRPerm(t *testing.T) {
	db := MakeDB(RandSource(), 256, 100)

	leftServer := NewPirPermServer(db)
	rightServer := NewPirPermServer(db)
	client := NewPIRClient(
		NewPirPermClient(RandSource()),
		RandSource(),
		[2]PirServer{leftServer, rightServer})

	assert.NilError(t, client.Init())
	const readIndex = 2
	val, err := client.Read(readIndex)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db[readIndex])
}
