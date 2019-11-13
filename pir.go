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
	Hint(hintReq io.Reader) (io.Reader, error)
	Answer(q io.Reader) (io.Reader, error)
}

// PIRClient is the interface that wraps the client methods.
type PIRClient interface {
	RequestHint() (io.Reader, error)
	InitHint(hint io.Reader) error
	Query(i int) ([]io.Reader, error)
	Reconstruct(answers []io.Reader) (string, error)
}
