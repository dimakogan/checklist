package main

import (
	"io"
)

// Database is an array-like abstraction for a database.
type Database interface {
	Length() int
	Get(i int) string
}

// PIRServer is the interface that wraps the server methods.
type PIRServer interface {
	Hint(hintReq io.Reader, hintWriter io.Writer) error
	Answer(q io.Reader, answerWriter io.Writer) error
}

// PIRClient is the interface that wraps the client methods.
type PIRClient interface {
	RequestHint(reqWriter io.Writer) error
	InitHint(hint io.Reader) error
	Query(i int, reqWriters []io.Writer) error
	Reconstruct(answers []io.Reader) (string, error)
}
