package boosted

import (
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

func MakeDB(nRows int, rowLen int) []Row {
	db := make([]Row, nRows)
	src := RandSource()
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
		db[i][0] = byte(i % 256)
		db[i][1] = 'A' + byte(i%256)
	}
	return db
}

func MakeKeys(nRows int) []uint32 {
	keys := make([]uint32, nRows)
	src := RandSource()
	for i := range keys {
		keys[i] = uint32(src.Int31()<<32) + uint32(i)
	}
	return keys
}

type DBDimensions struct {
	NumRecords int
	RecordSize int
}

func MakeDBWithDimensions(dim DBDimensions) []Row {
	return MakeDB(dim.NumRecords, dim.RecordSize)
}

type RecordIndexVal struct {
	Index int
	Key   uint32
	Value Row
}
