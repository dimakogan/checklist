package main

import (
	"bytes"
	"io"
	"testing"

	"gotest.tools/assert"
)

var data = []string{"A", "B", "C", "D"}

func testBasicRead(t *testing.T, client PIRClient, server PIRServer) {
	request := new(bytes.Buffer)
	response := new(bytes.Buffer)

	assert.NilError(t, client.RequestHint(request))
	assert.NilError(t, server.Hint(request, response))
	assert.NilError(t, client.InitHint(response))

	request.Reset()
	response.Reset()

	assert.NilError(t, client.Query(2, []io.Writer{request}))
	assert.NilError(t, server.Answer(request, response))
	val, err := client.Reconstruct([]io.Reader{response})
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
