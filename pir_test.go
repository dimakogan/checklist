package main

import (
	"testing"

	"gotest.tools/assert"
)

var db = []string{"A", "B", "C", "D"}

func testBasicRead(t *testing.T, client PIRClient, server PIRServer) {
	hintReq, err := client.RequestHint()
	assert.NilError(t, err)
	hintResp, err := server.Hint(hintReq)
	assert.NilError(t, err)
	assert.NilError(t, client.InitHint(hintResp))

	const readIndex = 2
	queryReq, err := client.Query(readIndex)
	assert.NilError(t, err)

	queryResp, err := server.Answer(queryReq[0])
	assert.NilError(t, err)
	val, err := client.Reconstruct([]*QueryResp{queryResp})
	assert.NilError(t, err)
	assert.Equal(t, val, db[readIndex])
}

// TestPIRStub is a simple end-to-end test.
func TestPIRStub(t *testing.T) {
	client := newPirClientStub()
	server := newPirServerStub(db)

	testBasicRead(t, client, server)
}
