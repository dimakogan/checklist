package main

import (
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

func (s pirServerStub) Hint(hintReq io.Reader, hintWriter io.Writer) error {
	return nil
}

func (s pirServerStub) Answer(q io.Reader, answerWriter io.Writer) error {
	var i uint32
	err := binary.Read(q, binary.LittleEndian, &i)
	if err != nil {
		return fmt.Errorf("Failed to parse query: %w", err)
	}
	_, err = answerWriter.Write([]byte(s.db.Get(int(i))))
	return err
}

type pirClientStub struct {
}

func newPirClientStub() PIRClient {
	return pirClientStub{}
}

func (c pirClientStub) RequestHint(reqWriter io.Writer) error {
	return nil
}

func (c pirClientStub) InitHint(hint io.Reader) error {
	return nil
}

func (c pirClientStub) Query(i int, reqWriters []io.Writer) error {
	if len(reqWriters) < 1 {
		return fmt.Errorf("Missing request writer")
	}
	binary.Write(reqWriters[0], binary.LittleEndian, uint32(i))
	return nil
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
