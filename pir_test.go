package main

import (
	"io"
	"testing"

	"gotest.tools/assert"
)

var data = []string{"A", "B", "C", "D"}

func testBasicRead(t *testing.T, client PIRClient, server PIRServer) {
	hintReq, err := client.RequestHint()
	assert.NilError(t, err)

	hint, err := server.Hint(hintReq)
	assert.NilError(t, err)

	assert.NilError(t, client.InitHint(hint))

	queries, err := client.Query(2)
	assert.NilError(t, err)
	assert.Equal(t, len(queries), 1)

	ans, err := server.Answer(queries[0])
	assert.NilError(t, err)

	val, err := client.Reconstruct([]io.Reader{ans})
	assert.NilError(t, err)
	assert.Equal(t, val, "C")
}

// TestPIRStub is a simple end-to-end test.
func TestPIRStub(t *testing.T) {
	client := newPirClientStub()
	db := databaseStub{data}
	server := newPirServerStub(db)

	testBasicRead(t, client, server)
}
