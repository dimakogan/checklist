package pir

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

func MakeRows(src *rand.Rand, nRows, rowLen int) []Row {
	db := make([]Row, nRows)
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
		db[i][0] = byte(i % 256)
		db[i][1] = 'A' + byte(i%256)
	}
	return db
}

func MakeDB(nRows int, rowLen int) StaticDB {
	return *StaticDBFromRows(MakeRows(RandSource(), nRows, rowLen))
}

func MakeKeys(src *rand.Rand, nRows int) []uint32 {
	keys := make([]uint32, nRows)
	for i := range keys {
		keys[i] = uint32(src.Int31()<<4) + uint32(i)
	}
	return keys
}

func MakeKeysRows(numRows, rowLen int) ([]uint32, []Row) {
	return MakeKeys(RandSource(), numRows), MakeRows(RandSource(), numRows, rowLen)
}
