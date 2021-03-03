package boosted

import "fmt"

type staticDB struct {
	numRows int
	rowLen  int
	flatDb  []byte
}

func (db *staticDB) GetRow(idx int, row *RowIndexVal) error {
	if idx < 0 || idx >= db.numRows {
		return fmt.Errorf("Index %d out of bounds [0,%d)", idx, db.numRows)
	}
	if idx == -1 {
		// return random row
		idx = RandSource().Int() % db.numRows
	}
	row.Value = db.flatDb[idx*db.rowLen : (idx+1)*db.rowLen]
	row.Index = idx
	row.Key = uint32(idx)
	return nil
}

func (db *staticDB) Slice(start, end int) []byte {
	return db.flatDb[start*db.rowLen : end*db.rowLen]
}

func (db *staticDB) Row(i int) Row {
	if i >= db.numRows {
		return nil
	}
	return Row(db.Slice(i, i+1))
}
