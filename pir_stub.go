package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

type databaseStub struct {
	data []string
}

func (db databaseStub) Length() int {
	return len(db.data)
}

func (db databaseStub) Get(i int) string {
	return db.data[i]
}

type pirServerStub struct {
	db Database
}

func newPirServerStub(db Database) PIRServer {
	return pirServerStub{db: db}
}

func (s pirServerStub) Hint(hintReq io.Reader) (io.Reader, error) {
	return &bytes.Buffer{}, nil
}

func (s pirServerStub) Answer(q io.Reader) (io.Reader, error) {
	var i uint32
	err := binary.Read(q, binary.LittleEndian, &i)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse query: %w", err)
	}
	return bytes.NewBufferString(s.db.Get(int(i))), nil
}

type pirClientStub struct {
}

func newPirClientStub() PIRClient {
	return pirClientStub{}
}

func (c pirClientStub) RequestHint() (io.Reader, error) {
	return bytes.NewBufferString(""), nil
}

func (c pirClientStub) InitHint(hint io.Reader) error {
	return nil
}

func (c pirClientStub) Query(i int) ([]io.Reader, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(i))
	return []io.Reader{buf}, nil
}

func (c pirClientStub) Reconstruct(answers []io.Reader) (string, error) {
	if len(answers) != 1 {
		return "", fmt.Errorf("Unexpected number of answers: have: %d, want: 1", len(answers))
	}
	b, err := ioutil.ReadAll(answers[0])
	if err != nil {
		return "", fmt.Errorf("Failed to read value from answer: %w", err)
	}
	return string(b), nil
}
