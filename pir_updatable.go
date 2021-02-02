package boosted

import (
	"fmt"
	"log"
	"math/rand"
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

	s.ops = append(s.ops, ops...)
	s.numRows += len(rows)
	s.flatDb = append(s.flatDb, flattenDb(rows)...)
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

var NumLayerActivations map[int]int
var NumLayerHintBytes map[int]int

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

	resp.Keys, resp.IsDeletion = opsToKeyUpdates(s.ops[firstPos:])
	resp.InitialTimestamp = int32(s.initialTimestamp + firstPos)
	return nil
}

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	clientSrc := rand.New(rand.NewSource(req.RandSeed))

	if int(req.DefragTimestamp) < s.defragTimestamp {
		return fmt.Errorf("Defragged since last key updates")
	}

	resp.BatchResps = make([]HintResp, len(req.Layers))
	for l, layer := range req.Layers {
		if layer.PirType != None {
			layerFlatDb := s.flatDb[layer.FirstRow*s.rowLen : (layer.FirstRow+layer.NumRows)*s.rowLen]
			pir := NewPirServerByType(layer.PirType, s.randSource, layerFlatDb, layer.NumRows, s.rowLen)
			pir.Hint(HintReq{RandSeed: int64(clientSrc.Uint64())}, &resp.BatchResps[l])
			resp.BatchResps[l].PirType = layer.PirType
		}
	}

	return nil
}

func opsToKeyUpdates(ops []dbOp) (Keys []uint32, IsDeletion []byte) {
	keys := make([]uint32, len(ops))
	isDeletion := make([]uint8, (len(keys)-1)/8+1)
	for j := range keys {
		keys[j] = ops[j].Key
		if ops[j].Delete {
			isDeletion[j/8] |= (1 << (j % 8))
		}
	}
	return keys, isDeletion
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
	if err := c.updateKeys(); err != nil {
		return err
	}
	return c.updateHint()
}

func (c *pirClientUpdatable) updateKeys() error {
	nextTimestamp := c.nextTimestamp()
	keyReq := KeyUpdatesReq{
		DefragTimestamp: int32(c.defragTimestamp),
		NextTimestamp:   int32(nextTimestamp),
	}
	var keyResp KeyUpdatesResp
	if err := c.servers[Left].KeyUpdates(keyReq, &keyResp); err != nil {
		return err
	}
	c.rowLen = keyResp.RowLen
	if keyResp.ShouldDeleteHistory {
		c.ops = []dbOp{}
		c.numRows = 0
		c.initialTimestamp = int(keyResp.InitialTimestamp)
		c.defragTimestamp = keyResp.DefragTimestamp
		c.keyToPos = make(map[uint32]int32)
		c.layers = c.freshLayers(0)
	}

	newOps := make([]dbOp, len(keyResp.Keys))
	for i := range keyResp.Keys {
		isDelete := (keyResp.IsDeletion[i/8] & (1 << (i % 8))) != 0
		newOps[i] = dbOp{
			Key:    keyResp.Keys[i],
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

	c.updatePositionMap(len(c.ops) - len(newOps))

	c.updateLayers(newOps)

	return nil
}

func smallestLayerSize(nRows int) int {
	return 10 * (*SecParam) * (*SecParam)
}

func (c pirClientUpdatable) LayersMaxSize(nRows int) []int {
	// if nRows == 0 {
	// 	return []int{}
	// }
	if c.pirType != Punc {
		return []int{nRows}
	}
	maxSize := []int{nRows}
	smallest := smallestLayerSize(nRows)
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

	initDebugCounters(len(layers))
	return layers
}

func initDebugCounters(numLayers int) {
	if NumLayerActivations == nil {
		NumLayerActivations = make(map[int]int)
		NumLayerHintBytes = make(map[int]int)
	}
	for i := 0; i < numLayers; i++ {
		if _, ok := NumLayerActivations[i]; !ok {
			NumLayerActivations[i] = 0
		}
		if _, ok := NumLayerHintBytes[i]; !ok {
			NumLayerHintBytes[i] = 0
		}
	}
}

func (c *pirClientUpdatable) updateLayers(ops []dbOp) {
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
		return
	}
	c.layers[i].numOps = numNewOps
	c.layers[i].numRows = numNewRows
	c.layers[i].firstRow = c.numRows - numNewRows
	c.layers[i].pir = nil
	return
}

func (c *pirClientUpdatable) updateHint() error {
	hintReq := HintReq{
		Layers:          make([]HintLayer, len(c.layers)),
		DefragTimestamp: int32(c.defragTimestamp),
		RandSeed:        int64(c.randSource.Uint64())}
	hintResp := HintResp{
		BatchResps: make([]HintResp, len(hintReq.Layers)),
	}

	needServer := false
	for l, layer := range c.layers {
		if layer.numRows != 0 && layer.pir == nil {
			if c.pirType != Punc || l == len(hintReq.Layers)-1 {
				hintResp.BatchResps[l].NumRows = layer.numRows
				hintResp.BatchResps[l].RowLen = c.rowLen
				hintResp.BatchResps[l].PirType = c.pirType
				if c.pirType == Punc && l == len(hintReq.Layers)-1 {
					hintResp.BatchResps[l].PirType = Matrix
				}
			} else {
				hintReq.Layers[l].FirstRow = layer.firstRow
				hintReq.Layers[l].NumRows = layer.numRows
				hintReq.Layers[l].PirType = Punc
				needServer = true
			}
		}
	}

	if needServer {
		if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
			return err
		}
	}

	if err := c.initHints(hintResp); err != nil {
		return err
	}

	return nil
}

func (c *pirClientUpdatable) initHints(resp HintResp) error {
	for l, subResp := range resp.BatchResps {
		if subResp.PirType == None {
			continue
		}
		c.layers[l].pir = NewPirClientByType(subResp.PirType, c.randSource)
		c.layers[l].pirType = subResp.PirType
		err := c.layers[l].pir.initHint(&subResp)
		if err != nil {
			return err
		}
		// Debug
		offlineBytes, err := SerializedSizeOf(subResp)
		if err != nil {
			return err
		}
		c.layers[l].hintNumBytes = offlineBytes
		NumLayerHintBytes[l] += len(resp.BatchResps[l].Hints) * resp.BatchResps[l].NumRowsPerBlock * resp.BatchResps[l].RowLen
	}
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

func (c pirClientUpdatable) Read(key uint32) (Row, error) {
	queryReq, reconstructFunc := c.query(int(key))
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %x", key)
	}
	responses := make([]QueryResp, 2)
	errs := make([]error, 2)
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
		{
			LatestKeyTimestamp: int32(c.nextTimestamp()),
			BatchReqs:          make([]QueryReq, len(c.layers))},
		{
			LatestKeyTimestamp: int32(c.nextTimestamp()),
			BatchReqs:          make([]QueryReq, len(c.layers))},
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
			matchingLayer = l
		} else {
			q = layer.pir.dummyQuery()
		}
		layerEnd += layer.numRows
		req[Left].BatchReqs[l] = q[Left]
		req[Right].BatchReqs[l] = q[Right]
		req[Left].BatchReqs[l].FirstRow = int32(c.layers[l].firstRow)
		req[Right].BatchReqs[l].FirstRow = int32(c.layers[l].firstRow)
		req[Left].BatchReqs[l].NumRows = int32(c.layers[l].numRows)
		req[Right].BatchReqs[l].NumRows = int32(c.layers[l].numRows)
		req[Left].BatchReqs[l].PirType = c.layers[l].pirType
		req[Right].BatchReqs[l].PirType = c.layers[l].pirType
	}
	return req, func(resps []QueryResp) (Row, error) {
		row, err := reconstructFunc([]QueryResp{
			resps[Left].BatchResps[matchingLayer],
			resps[Right].BatchResps[matchingLayer]})
		return row, err
	}
}

func (c *pirClientUpdatable) StorageNumBytes() int {
	numBytes := 0

	keyUpdates, isDeletion := opsToKeyUpdates(c.ops)
	keysBytes, err := SerializedSizeOf(keyUpdates)
	if err != nil {
		return 0
	}
	numBytes += keysBytes
	numBytes += len(isDeletion)

	for _, l := range c.layers {
		numBytes += l.hintNumBytes
	}

	return numBytes
}
