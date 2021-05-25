package pir

import (
	"fmt"

	"github.com/lukechampine/fastxor"
)

type StaticDB struct {
	NumRows int
	RowLen  int
	FlatDb  []byte
}

func (db *StaticDB) Slice(start, end int) []byte {
	return db.FlatDb[start*db.RowLen : end*db.RowLen]
}

func (db *StaticDB) Row(i int) Row {
	if i >= db.NumRows {
		return nil
	}
	return Row(db.Slice(i, i+1))
}

func StaticDBFromRows(data []Row) *StaticDB {
	if len(data) < 1 {
		return &StaticDB{0, 0, nil}
	}

	rowLen := len(data[0])
	flatDb := make([]byte, rowLen*len(data))

	for i, v := range data {
		if len(v) != rowLen {
			fmt.Printf("Got row[%v] %v %v\n", i, len(v), rowLen)
			panic("Database rows must all be of the same length")
		}

		copy(flatDb[i*rowLen:], v[:])
	}
	return &StaticDB{len(data), rowLen, flatDb}
}

func (db StaticDB) Hint(req HintReq, resp *HintResp) (err error) {
	*resp, err = req.Process(db)
	return err
}
func (db StaticDB) Answer(q QueryReq, resp *interface{}) (err error) {
	*resp, err = q.Process(db)
	return err
}

type DBParams struct {
	NRows  int
	RowLen int
}

func (p *DBParams) NumRows() int {
	return p.NRows
}

func (db StaticDB) Params() *DBParams {
	return &DBParams{db.NumRows, db.RowLen}
}

func xorInto(a []byte, b []byte) {
	if len(a) != len(b) {
		panic("Tried to XOR byte-slices of unequal length.")
	}

	fastxor.Bytes(a, a, b)

	// for i := 0; i < len(a); i++ {
	// 	a[i] = a[i] ^ b[i]
	// }
}
