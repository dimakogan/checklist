package boosted

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
)

func RandSource() *rand.Rand {
	return rand.New(rand.NewSource(17))
}

func MasterKey() []byte {
	key := make([]byte, 16)
	io.ReadFull(RandSource(), key)
	return key
}

func MakeDB(src *rand.Rand, nRows int, rowLen int) []Row {
	db := make([]Row, nRows)
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
		db[i][0] = byte(i % 256)
		db[i][1] = 'A' + byte(i%256)
	}
	return db
}

func MakeKeys(src *rand.Rand, nRows int) []uint32 {
	keys := make([]uint32, nRows)
	for i := range keys {
		keys[i] = uint32(src.Int31()<<4) + uint32(i)
	}
	return keys
}

type TestConfig struct {
	NumRows int
	RowLen  int

	PirType   PirType
	Updatable bool

	UpdateSize int

	PresetRows []RowIndexVal

	RandSeed int64
}

func (c TestConfig) String() string {
	return fmt.Sprintf("%s/n=%d,r=%d", c.PirType, c.NumRows, c.RowLen)
}

type RowIndexVal struct {
	Index int
	Key   uint32
	Value Row
}

func SerializedSizeOf(e interface{}) (int, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(e); err != nil {
		return -1, fmt.Errorf("Failed to Encode: %s", err)
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
