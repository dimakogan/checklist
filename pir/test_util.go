package pir

import (
	"math/rand"
)

var masterKey PRGKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 'A', 'B', 'C', 'D', 'E', 'F'}
var randReader *rand.Rand = rand.New(NewBufPRG(NewPRG(&masterKey)))

func RandSource() *rand.Rand {
	//return rand.New(rand.NewSource(17))
	return randReader
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
		keys[i] = uint32(src.Int31())
	}
	return keys
}

func MakeKeysRows(numRows, rowLen int) ([]uint32, []Row) {
	return MakeKeys(RandSource(), numRows), MakeRows(RandSource(), numRows, rowLen)
}
