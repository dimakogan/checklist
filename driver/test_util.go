package driver

import (
	"bytes"
	"fmt"

	"checklist/rpc"

	"github.com/ugorji/go/codec"
)

type TestConfig struct {
	NumRows int
	RowLen  int

	Updatable bool

	UpdateSize int

	PresetRows []RowIndexVal

	// Seed used to generate random data in database. Not used for cryptographic operations.
	DataRandSeed int64

	MeasureBandwidth bool
}

func (c TestConfig) String() string {
	return fmt.Sprintf("n=%d,r=%d", c.NumRows, c.RowLen)
}

func SerializedSizeOf(e interface{}) (int, error) {
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, rpc.CodecHandle((RegisteredTypes())))
	err := enc.Encode(e)
	if err != nil {
		panic(err)
	}
	return buf.Len(), nil
}

// Disgusting hack since testing.Benchmark hides all logs and failures
type ErrorPrinter struct {
}

func (ep ErrorPrinter) Log(args ...interface{}) {
	fmt.Println(args...)
}

func (ep ErrorPrinter) FailNow() {
	panic("Assertion failed")
}

func (ep ErrorPrinter) Fail() {
	panic("Assertion failed")
}
