package boosted

import "math/rand"

func RandSource() *rand.Rand {
	return rand.New(rand.NewSource(17))
}

func MakeDB(nRows int, rowLen int) []Row {
	db := make([]Row, nRows)
	src := RandSource()
	for i := range db {
		db[i] = make([]byte, rowLen)
		src.Read(db[i])
	}
	return db
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
	Value Row
}
