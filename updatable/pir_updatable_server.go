package updatable

import (
	"fmt"
	"log"
	"sort"

	"checklist/pir"

	"github.com/elliotchance/orderedmap"
)

type dbOp struct {
	Key    uint32
	Delete bool
	data   pir.Row
}

type Server struct {
	pir.StaticDB
	initialTimestamp int
	defragTimestamp  int
	ops              []dbOp
	kv               *orderedmap.OrderedMap

	curTimestamp int32

	defragRatio float64
}

func NewUpdatableServer() *Server {
	s := Server{
		curTimestamp: 0,
		defragRatio:  4,
		kv:           orderedmap.NewOrderedMap(),
	}
	return &s
}

func (s *Server) Row(idx int) (uint32, pir.Row, error) {
	for e, pos := s.kv.Front(), 0; e != nil; e, pos = e.Next(), pos+1 {
		if pos == idx {
			return uint32(e.Key.(uint32)), pir.Row(e.Value.(pir.Row)), nil
		}
	}
	return 0, nil, fmt.Errorf("Index %d out of bounds [0:%d)", idx, s.kv.Len())
}

func (s *Server) NumKeys() int {
	return s.kv.Len()
}

func (s *Server) SomeKeys(num int) []uint32 {
	keys := make([]uint32, num)
	for e, pos := s.kv.Front(), 0; e != nil; e, pos = e.Next(), pos+1 {
		if pos == num {
			break
		}
		keys[pos] = uint32(e.Key.(uint32))
	}

	return keys
}

func (s *Server) AddRows(keys []uint32, rows []pir.Row) {
	if len(rows) == 0 {
		return
	}
	if s.RowLen == 0 {
		s.RowLen = len(rows[0])
	} else if s.RowLen != len(rows[0]) {
		log.Fatalf("Different row length added, expected: %d, got: %d", s.RowLen, len(rows[0]))
	}

	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], data: rows[i]}
	}

	sort.Slice(ops, func(i, j int) bool { return ops[i].Key < ops[j].Key })

	for i := range ops {
		s.kv.Set(ops[i].Key, ops[i].data)
	}

	s.ops = append(s.ops, ops...)
	s.NumRows += len(rows)
	s.FlatDb = append(s.FlatDb, opsToFlatDB(ops)...)
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

func (s *Server) DeleteRows(keys []uint32) {
	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], Delete: true}
		s.kv.Delete(keys[i])
	}
	s.ops = append(s.ops, ops...)

	if len(s.ops) > int(s.defragRatio*float64(s.kv.Len())) {
		endDefrag := len(s.ops) / 2
		newOps, _ := defrag(s.ops, endDefrag)
		s.defragTimestamp = s.initialTimestamp + endDefrag
		s.initialTimestamp += (len(s.ops) - len(newOps))
		s.ops = newOps
		s.StaticDB = *pir.StaticDBFromRows(opsToRows(s.ops))
	}
}

func (s *Server) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	resp.DefragTimestamp = s.defragTimestamp
	resp.RowLen = s.RowLen

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

func opsToRows(ops []dbOp) []pir.Row {
	db := make([]pir.Row, 0, len(ops))
	for _, op := range ops {
		if !op.Delete {
			db = append(db, op.data)
		}
	}
	return db
}
