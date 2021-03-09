package boosted

import (
	"fmt"
	"log"
	"sort"

	"github.com/elliotchance/orderedmap"
)

type dbOp struct {
	Key    uint32
	Delete bool
	data   Row
}

type pirUpdatableServer struct {
	staticDB
	initialTimestamp int
	defragTimestamp  int
	ops              []dbOp
	kv               *orderedmap.OrderedMap

	curTimestamp int32

	defragRatio float64
}

func NewPirUpdatableServer() *pirUpdatableServer {
	s := pirUpdatableServer{
		curTimestamp: 0,
		defragRatio:  4,
		kv:           orderedmap.NewOrderedMap(),
	}
	return &s
}

func (s *pirUpdatableServer) Hint(req HintReq, resp *HintResp) error {
	// if int(req.DefragTimestamp) < s.defragTimestamp {
	// 	return fmt.Errorf("Defragged since last key updates")
	// }

	if req.PirType != None {
		layerFlatDb := s.Slice(req.FirstRow, req.FirstRow+req.NumRows)
		pir := NewPirServerByType(req.PirType, &staticDB{req.NumRows, s.rowLen, layerFlatDb})
		pir.Hint(req, resp)
		resp.PirType = req.PirType
	}

	return nil
}

func (s *pirUpdatableServer) Answer(req QueryReq, resp *QueryResp) error {
	resp.BatchResps = make([]QueryResp, len(req.BatchReqs))
	for l, q := range req.BatchReqs {
		if q.NumRows == 0 {
			continue
		}
		layerFlatDb := s.Slice(int(q.FirstRow), int(q.FirstRow+q.NumRows))
		pir := NewPirServerByType(q.PirType, &staticDB{int(q.NumRows), s.rowLen, layerFlatDb})
		err := pir.Answer(q, &(resp.BatchResps[l]))

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *pirUpdatableServer) NumRows() int {
	return s.kv.Len()
}

func (s *pirUpdatableServer) GetRow(idx int, out *RowIndexVal) error {
	if idx == -1 {
		// return random row
		idx = RandSource().Int() % s.kv.Len()
	}

	for e, pos := s.kv.Front(), 0; e != nil; e, pos = e.Next(), pos+1 {
		if pos == idx {
			out.Key = uint32(e.Key.(uint32))
			out.Value = Row(e.Value.(Row))
			out.Index = idx
			return nil
		}
	}
	return fmt.Errorf("Index %d out of bounds [0:%d)", idx, s.numRows)
}

func (s *pirUpdatableServer) SomeKeys(num int) []uint32 {
	keys := make([]uint32, num)
	for e, pos := s.kv.Front(), 0; e != nil; e, pos = e.Next(), pos+1 {
		if pos == num {
			break
		}
		keys[pos] = uint32(e.Key.(uint32))
	}

	return keys
}

func (s *pirUpdatableServer) AddRows(keys []uint32, rows []Row) {
	if len(rows) == 0 {
		return
	}
	if s.rowLen == 0 {
		s.rowLen = len(rows[0])
	} else if s.rowLen != len(rows[0]) {
		log.Fatalf("Different row length added, expected: %d, got: %d", s.rowLen, len(rows[0]))
	}

	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], data: rows[i]}
		s.kv.Set(keys[i], rows[i])
	}

	sort.Slice(ops, func(i, j int) bool { return ops[i].Key < ops[j].Key })

	s.ops = append(s.ops, ops...)
	s.numRows += len(rows)
	s.flatDb = append(s.flatDb, opsToFlatDB(ops)...)
}

func opsToFlatDB(ops []dbOp) []byte {
	rowLen := len(ops[0].data)
	flatDb := make([]byte, rowLen*len(ops))

	for i, v := range ops {
		if len(v.data) != rowLen {
			fmt.Printf("Got row[%v] %v %v\n", i, len(v.data), rowLen)
			panic("Database rows must all be of the same length")
		}
		copy(flatDb[i*rowLen:], v.data[:])
	}
	return flatDb
}

func (s *pirUpdatableServer) DeleteRows(keys []uint32) {
	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], Delete: true}
		s.kv.Delete(keys[i])
	}
	s.ops = append(s.ops, ops...)

	if len(s.ops) > int(s.defragRatio*float64(s.kv.Len())) {
		endDefrag := len(s.ops) / 2
		newOps, numRemoved := defrag(s.ops, endDefrag)
		s.defragTimestamp = s.initialTimestamp + endDefrag
		s.initialTimestamp += (len(s.ops) - len(newOps))
		s.ops = newOps
		s.flatDb = flattenDb(opsToRows(s.ops))
		s.numRows -= numRemoved
	}
}

func (s *pirUpdatableServer) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	resp.DefragTimestamp = s.defragTimestamp
	resp.RowLen = s.rowLen

	nextTimestamp := int(req.NextTimestamp)
	if nextTimestamp < s.defragTimestamp {
		resp.ShouldDeleteHistory = true
		nextTimestamp = s.initialTimestamp
	}
	firstPos := 0

	if nextTimestamp >= s.initialTimestamp {
		firstPos = nextTimestamp - s.initialTimestamp
	}
	if firstPos == len(s.ops) {
		return nil
	}

	var err error
	if opsToKeyUpdates(s.ops[firstPos:], resp) != nil {
		return fmt.Errorf("Failed to convert ops to keys: %v", err)
	}
	resp.InitialTimestamp = int32(s.initialTimestamp + firstPos)
	return nil
}

func defrag(ops []dbOp, endDefrag int) (newOps []dbOp, numRemoved int) {
	freed := 0

	keyToPos := make(map[uint32]int)
	for i, op := range ops[0:endDefrag] {
		_, prevExists := keyToPos[op.Key]
		if prevExists {
			numRemoved++
			freed++
		}
		if op.Delete {
			freed++
			delete(keyToPos, op.Key)
		} else {
			keyToPos[op.Key] = i
		}
	}

	// Push forward all defragmented rows
	newOps = make([]dbOp, len(ops)-freed)
	newPos := 0
	for i, op := range ops[0:endDefrag] {
		if pos, ok := keyToPos[op.Key]; ok && pos == i {
			newOps[newPos] = op
			newPos++
		}
	}
	copy(newOps[newPos:], ops[endDefrag:])
	return newOps, numRemoved
}

func opsToKeyUpdates(ops []dbOp, keyUpdate *KeyUpdatesResp) error {
	keys := make([]uint32, len(ops))
	keyUpdate.IsDeletion = make([]uint8, (len(keys)-1)/8+1)
	hasDeletion := false
	for j := range keys {
		keys[j] = ops[j].Key
		if ops[j].Delete {
			keyUpdate.IsDeletion[j/8] |= (1 << (j % 8))
			hasDeletion = true
		}
	}
	if !hasDeletion {
		keyUpdate.IsDeletion = nil
	}
	if sort.SliceIsSorted(keys, func(i, j int) bool { return keys[i] < keys[j] }) {
		var err error
		keyUpdate.KeysRice, err = EncodeRiceIntegers(keys)
		return err
	} else {
		keyUpdate.Keys = keys
		return nil
	}
}

func opsToRows(ops []dbOp) []Row {
	db := make([]Row, 0, len(ops))
	for _, op := range ops {
		if !op.Delete {
			db = append(db, op.data)
		}
	}
	return db
}
