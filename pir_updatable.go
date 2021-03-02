package boosted

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"

	"github.com/elliotchance/orderedmap"
)

type dbOp struct {
	Key    uint32
	Delete bool
	data   Row
}

type pirServerUpdatable struct {
	initialTimestamp int
	defragTimestamp  int
	numRows          int
	rowLen           int
	ops              []dbOp
	flatDb           []byte
	kv               *orderedmap.OrderedMap

	pirType PirType

	randSource *rand.Rand

	curTimestamp int32

	defragRatio float64
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

func NewPirServerUpdatable(source *rand.Rand) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
		defragRatio:  4,
		kv:           orderedmap.NewOrderedMap(),
	}
	return &s
}

func (s *pirServerUpdatable) NumRows(none int, out *int) error {
	*out = s.kv.Len()
	return nil
}

func (s *pirServerUpdatable) GetRow(idx int, out *RowIndexVal) error {
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
	return fmt.Errorf("Index %d out of bounds", idx)
}

func (s *pirServerUpdatable) SomeKeys(num int) []uint32 {
	keys := make([]uint32, num)
	for e, pos := s.kv.Front(), 0; e != nil; e, pos = e.Next(), pos+1 {
		if pos == num {
			break
		}
		keys[pos] = uint32(e.Key.(uint32))
	}

	return keys
}

func (s *pirServerUpdatable) AddRows(keys []uint32, rows []Row) {
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

func (s *pirServerUpdatable) DeleteRows(keys []uint32) {
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

func (s pirServerUpdatable) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
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

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	if int(req.DefragTimestamp) < s.defragTimestamp {
		return fmt.Errorf("Defragged since last key updates")
	}

	if req.PirType != None {
		layerFlatDb := s.flatDb[req.FirstRow*s.rowLen : (req.FirstRow+req.NumRows)*s.rowLen]
		pir := NewPirServerByType(req.PirType, s.randSource, layerFlatDb, req.NumRows, s.rowLen)
		pir.Hint(req, resp)
		resp.PirType = req.PirType
	}

	return nil
}

func opsToKeyUpdates(ops []dbOp, keyUpdate *KeyUpdatesResp) error {
	keys := make([]uint32, len(ops))
	keyUpdate.IsDeletion = make([]uint8, (len(keys)-1)/8+1)
	for j := range keys {
		keys[j] = ops[j].Key
		if ops[j].Delete {
			keyUpdate.IsDeletion[j/8] |= (1 << (j % 8))
		}
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

func (s pirServerUpdatable) Answer(req QueryReq, resp *QueryResp) error {

	resp.BatchResps = make([]QueryResp, len(req.BatchReqs))
	for l, q := range req.BatchReqs {
		if q.NumRows == 0 {
			continue
		}
		// start := time.Now()
		layerFlatDb := s.flatDb[int(q.FirstRow)*s.rowLen : int(q.FirstRow+q.NumRows)*s.rowLen]
		pir := NewPirServerByType(q.PirType, s.randSource, layerFlatDb, int(q.NumRows), s.rowLen)
		err := pir.Answer(q, &(resp.BatchResps[l]))
		//		log.Printf("pirServerPunc::Answer layer: %d | nRows: %d | time: %dµs", l, s.layers[l].numRows, time.Since(start).Microseconds())

		if err != nil {
			return err
		}
	}
	return nil
}

type Layer struct {
	MaxSize int
	// Including both deletes and adds. This is not the same as len(db).
	NumOps int

	FirstRow int
	NumRows  int

	PirType PirType

	pir pirClientImpl

	// debug
	hintNumBytes int
}

type clientLayer struct {
	maxSize int

	firstRow int
	numRows  int
	numOps   int

	pirType PirType
	pir     pirClientImpl

	// debug
	hintNumBytes int
}

type pirClientUpdatable struct {
	pirType          PirType
	randSource       *rand.Rand
	initialTimestamp int
	defragTimestamp  int
	numRows          int
	rowLen           int
	ops              []dbOp
	keyToPos         map[uint32]int32
	layers           []clientLayer
	servers          [2]PirUpdatableServer

	// For testing
	smallestLayerSizeOverride int
	CallAsync                 bool
}

func NewPirClientUpdatable(source *rand.Rand, pirType PirType, servers [2]PirUpdatableServer) *pirClientUpdatable {
	return &pirClientUpdatable{
		randSource: source,
		pirType:    pirType,
		servers:    servers,
		keyToPos:   make(map[uint32]int32)}
}

func (c *pirClientUpdatable) Init() error {
	err := c.Update()
	return err
}

func (c *pirClientUpdatable) Keys() []uint32 {
	keys := make([]uint32, 0, len(c.keyToPos))
	for k := range c.keyToPos {
		keys = append(keys, k)
	}
	return keys
}

func (c *pirClientUpdatable) Update() error {
	nextTimestamp := c.nextTimestamp()
	keyReq := KeyUpdatesReq{
		DefragTimestamp: int32(c.defragTimestamp),
		NextTimestamp:   int32(nextTimestamp),
	}
	var keyResp KeyUpdatesResp
	if err := c.servers[Left].KeyUpdates(keyReq, &keyResp); err != nil {
		return err
	}

	var newOps []dbOp
	var err error
	if newOps, err = c.processKeyUpdate(&keyResp); err != nil {
		return err
	}
	c.updatePositionMap(len(c.ops) - len(newOps))
	latestLayer := c.updateLayers(newOps)
	return c.initLayerHint(latestLayer)
}

func (c *pirClientUpdatable) processKeyUpdate(keyResp *KeyUpdatesResp) ([]dbOp, error) {
	c.rowLen = keyResp.RowLen
	if keyResp.ShouldDeleteHistory {
		c.ops = []dbOp{}
		c.numRows = 0
		c.initialTimestamp = int(keyResp.InitialTimestamp)
		c.defragTimestamp = keyResp.DefragTimestamp
		c.keyToPos = make(map[uint32]int32)
		c.layers = c.freshLayers(0)
	}
	var keys []uint32
	if keyResp.KeysRice != nil {
		var err error
		keys, err = DecodeRiceIntegers(keyResp.KeysRice)
		if err != nil {
			return nil, fmt.Errorf("Failed to Rice-decode key updates: %v", err)
		}
	} else {
		keys = keyResp.Keys
	}

	newOps := make([]dbOp, len(keys))
	for i := range keys {
		isDelete := (keyResp.IsDeletion[i/8] & (1 << (i % 8))) != 0
		newOps[i] = dbOp{
			Key:    keys[i],
			Delete: isDelete,
		}
	}

	if keyResp.DefragTimestamp > c.defragTimestamp {
		defraggedOps, _ := defrag(c.ops, keyResp.DefragTimestamp-c.initialTimestamp)
		c.defragTimestamp = keyResp.DefragTimestamp
		c.initialTimestamp += (len(c.ops) - len(defraggedOps))
		c.ops = append(defraggedOps, newOps...)
		c.numRows = 0
		c.keyToPos = make(map[uint32]int32)
		c.layers = c.freshLayers(0)
		newOps = c.ops
	} else {
		c.ops = append(c.ops, newOps...)
	}
	return newOps, nil
}

func (c *pirClientUpdatable) smallestLayerSize(nRows int) int {
	if c.smallestLayerSizeOverride != 0 {
		return c.smallestLayerSizeOverride
	}
	return 10 * SecParam * SecParam
}

func (c *pirClientUpdatable) LayersMaxSize(nRows int) []int {
	// if nRows == 0 {
	// 	return []int{}
	// }
	if c.pirType != Punc {
		return []int{nRows}
	}
	maxSize := []int{nRows}
	smallest := c.smallestLayerSize(nRows)
	for maxSize[len(maxSize)-1] > smallest {
		maxSize = append(maxSize, maxSize[len(maxSize)-1]/2)
	}
	return maxSize
}

func (c *pirClientUpdatable) freshLayers(numRows int) []clientLayer {
	maxSizes := c.LayersMaxSize(numRows)

	layers := make([]clientLayer, len(maxSizes))
	if len(layers) == 0 {
		return layers
	}
	for i := range layers {
		layers[i].maxSize = maxSizes[i]
	}

	return layers
}

func (c *pirClientUpdatable) updateLayers(ops []dbOp) int {
	numNewRows := 0
	for _, op := range ops {
		if !op.Delete {
			numNewRows++
		}
	}
	numNewOps := len(ops)
	var i int
	for i = len(c.layers) - 1; i >= 0; i-- {
		numNewRows += c.layers[i].numRows
		numNewOps += c.layers[i].numOps
		if numNewRows <= c.layers[i].maxSize {
			break
		}

		c.layers[i] = clientLayer{maxSize: c.layers[i].maxSize}
	}
	if i <= 0 {
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		c.layers = c.freshLayers(numNewRows)
		i = 0
	}
	if len(c.layers) == 0 {
		return -1
	}
	layer := &c.layers[i]
	layer.numOps = numNewOps
	layer.numRows = numNewRows
	layer.firstRow = c.numRows - numNewRows
	layer.pir = nil
	return i
}

// Debug
var RequestedHintNumRows []int

func (c *pirClientUpdatable) initLayerHint(layerNum int) error {
	if layerNum < 0 || c.layers[layerNum].numRows == 0 {
		return nil
	}
	layer := &c.layers[layerNum]
	var hintResp HintResp
	if layerNum == len(c.layers)-1 {
		pirType := c.pirType
		if c.pirType == Punc {
			pirType = DPF
		}
		hintResp = HintResp{
			NumRows: layer.numRows,
			RowLen:  c.rowLen,
			PirType: pirType,
		}
	} else {
		hintReq := HintReq{
			RandSeed:        int64(c.randSource.Uint64()),
			DefragTimestamp: int32(c.defragTimestamp),
			FirstRow:        layer.firstRow,
			NumRows:         layer.numRows,
			PirType:         Punc,
		}

		// Debug
		if RequestedHintNumRows != nil {
			RequestedHintNumRows = append(RequestedHintNumRows, layer.numRows)
		}
		if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
			return err
		}
	}

	return c.initHint(layer, &hintResp)
}

func (c *pirClientUpdatable) initHint(layer *clientLayer, resp *HintResp) error {
	layer.pir = NewPirClientByType(resp.PirType, c.randSource)
	layer.pirType = resp.PirType
	err := layer.pir.initHint(resp)
	if err != nil {
		return err
	}
	// Debug
	offlineBytes, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	layer.hintNumBytes = offlineBytes
	return nil
}

func (c *pirClientUpdatable) updatePositionMap(fromOpNumber int) {
	for i := fromOpNumber; i < len(c.ops); i++ {
		op := c.ops[i]
		if op.Delete {
			// propagate deletes backwards to previous layers
			delete(c.keyToPos, op.Key)
			continue
		}
		c.keyToPos[op.Key] = int32(c.numRows)
		c.numRows++
	}
}

func (c *pirClientUpdatable) nextTimestamp() int {
	return c.initialTimestamp + len(c.ops)
}

func (c *pirClientUpdatable) Read(key uint32) (Row, error) {
	queryReq, reconstructFunc := c.query(int(key))
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %x", key)
	}
	responses := make([]QueryResp, 2)
	errs := make([]error, 2)

	if c.CallAsync {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			errs[Left] = c.servers[Left].Answer(queryReq[Left], &responses[Left])
		}()
		go func() {
			defer wg.Done()
			errs[Right] = c.servers[Right].Answer(queryReq[Right], &responses[Right])
		}()
		wg.Wait()
	} else {
		errs[Left] = c.servers[Left].Answer(queryReq[Left], &responses[Left])
		errs[Right] = c.servers[Right].Answer(queryReq[Right], &responses[Right])
	}
	if errs[Left] != nil {
		return nil, errs[Left]
	}
	if errs[Right] != nil {
		return nil, errs[Left]
	}

	return reconstructFunc(responses)
}

func (c *pirClientUpdatable) query(i int) ([]QueryReq, ReconstructFunc) {
	req := []QueryReq{
		{LatestKeyTimestamp: int32(c.nextTimestamp())},
		{LatestKeyTimestamp: int32(c.nextTimestamp())},
	}
	var reconstructFunc ReconstructFunc

	// Slight hack for now: using i as key
	pos, ok := c.keyToPos[uint32(i)]
	if !ok {
		return nil, nil
	}

	var layerEnd int
	matchingLayer := 0
	for l, layer := range c.layers {
		var q []QueryReq
		if layer.pir == nil {
			continue
		}
		if layerEnd <= int(pos) && int(pos) < layerEnd+layer.numRows {
			q, reconstructFunc = layer.pir.query(int(pos) - layerEnd)
			matchingLayer = len(req[Left].BatchReqs)
		} else {
			q = layer.pir.dummyQuery()
		}
		for s := range []int{Left, Right} {
			q[s].FirstRow = int32(c.layers[l].firstRow)
			q[s].NumRows = int32(c.layers[l].numRows)
			q[s].PirType = c.layers[l].pirType
			req[s].BatchReqs = append(req[s].BatchReqs, q[s])
		}
		layerEnd += layer.numRows
	}
	return req, func(resps []QueryResp) (Row, error) {
		row, err := reconstructFunc([]QueryResp{
			resps[Left].BatchResps[matchingLayer],
			resps[Right].BatchResps[matchingLayer]})
		return row, err
	}
}

type compressedMap struct {
	Present []byte
	Keys    []int32
}

func compressPosMap(k2v map[uint32]int32, valRange int) compressedMap {
	v2k := make([]int32, valRange)
	for k, v := range k2v {
		v2k[v] = int32(k)
	}
	present := make([]byte, valRange)
	keys := make([]int32, 0, len(k2v))
	for v, k := range v2k {
		keys = append(keys, k)
		present[v/8] |= (1 << (v % 8))
	}
	return compressedMap{present, keys}
}

func (c *pirClientUpdatable) StorageNumBytes() int {
	numBytes := 0

	keysBytes, err := SerializedSizeOf(compressPosMap(c.keyToPos, c.numRows))
	if err != nil {
		log.Fatalf("%s", err)
		return 0
	}
	numBytes += keysBytes

	for _, l := range c.layers {
		numBytes += l.hintNumBytes
	}

	return numBytes
}

func (c *pirClientUpdatable) NumLayers() int {
	return len(c.layers)
}
