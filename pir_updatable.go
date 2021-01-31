package boosted

import (
	"fmt"
	"log"
	"math/rand"

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

func smallestLayerSize(nRows int) int {
	return 10 * (*SecParam) * (*SecParam)
}

func (s pirServerUpdatable) layersMaxSize(nRows int) []int {
	// if nRows == 0 {
	// 	return []int{}
	// }
	if s.pirType != Punc {
		return []int{nRows}
	}
	maxSize := []int{nRows}
	smallest := smallestLayerSize(nRows)
	for maxSize[len(maxSize)-1] > smallest {
		maxSize = append(maxSize, maxSize[len(maxSize)-1]/2)
	}
	return maxSize
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

func (s *pirServerUpdatable) initLayers(nRows int) []Layer {
	maxSizes := s.layersMaxSize(nRows)

	layers := make([]Layer, len(maxSizes))
	if len(layers) == 0 {
		return layers
	}
	for i := range layers {
		layers[i].MaxSize = maxSizes[i]
		layers[i].PirType = s.pirType
	}

	// Even when using PirPunc, the smallest layer always uses matrix
	if s.pirType == Punc {
		layers[len(layers)-1].PirType = Matrix
	}

	if NumLayerActivations == nil {
		NumLayerActivations = make(map[int]int)
		NumLayerHintBytes = make(map[int]int)
	}
	for i := range layers {
		if _, ok := NumLayerActivations[i]; !ok {
			NumLayerActivations[i] = 0
		}
		if _, ok := NumLayerHintBytes[i]; !ok {
			NumLayerHintBytes[i] = 0
		}
	}

	return layers
}

func NewPirServerUpdatable(source *rand.Rand, pirType PirType) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
		defragRatio:  4,
		pirType:      pirType,
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

func (s pirServerUpdatable) updateLayers(layers []Layer, ops []dbOp) []Layer {
	numNewRows := 0
	for _, op := range ops {
		if !op.Delete {
			numNewRows++
		}
	}
	numNewOps := len(ops)
	var i int
	for i = len(layers) - 1; i >= 0; i-- {
		numNewRows += layers[i].NumRows
		numNewOps += layers[i].NumOps
		if numNewRows <= layers[i].MaxSize {
			break
		}

		layers[i].NumOps = 0
		layers[i].NumRows = 0
		layers[i].FirstRow = 0
	}
	if i <= 0 {
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		layers = s.initLayers(numNewRows)
		i = 0
	}
	if len(layers) == 0 {
		return layers
	}
	layers[i].NumOps = numNewOps
	layers[i].NumRows = numNewRows
	layers[i].FirstRow = s.numRows - numNewRows
	return layers
}

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	clientSrc := rand.New(rand.NewSource(req.RandSeed))
	resp.DefragTimestamp = s.defragTimestamp
	if int(req.NextTimestamp) < s.defragTimestamp {
		resp.ShouldDeleteHistory = true
		req.NextTimestamp = int32(s.initialTimestamp)
	}
	resp.KeyUpdates = s.returnDiffKeys(int(req.NextTimestamp))

	if int(req.DefragTimestamp) < s.defragTimestamp {
		resp.Layers = s.updateLayers([]Layer{}, s.ops)
	} else {
		newOps := s.ops[len(s.ops)-len(resp.KeyUpdates.Keys):]
		resp.Layers = s.updateLayers(req.Layers, newOps)
	}

	layerEnd := s.initialTimestamp
	resp.BatchResps = make([]HintResp, len(resp.Layers))
	for l, layer := range resp.Layers {
		layerEnd += layer.NumOps
		if layer.NumRows != 0 && (int(req.NextTimestamp) < layerEnd) {
			layerFlatDb := s.flatDb[layer.FirstRow*s.rowLen : (layer.FirstRow+layer.NumRows)*s.rowLen]
			pir := NewPirServerByType(layer.PirType, s.randSource, layerFlatDb, layer.NumRows, s.rowLen)
			pir.Hint(
				HintReq{RandSeed: int64(clientSrc.Uint64())},
				&resp.BatchResps[l])
			resp.BatchResps[l].PirType = layer.PirType
			NumLayerActivations[l]++
			NumLayerHintBytes[l] += len(resp.BatchResps[l].Hints) * resp.BatchResps[l].NumRowsPerBlock * resp.BatchResps[l].RowLen
		}
		resp.BatchResps[l].NumRows = layer.NumRows
	}

	return nil
}

func opsToKeyUpdates(ops []dbOp) KeyUpdates {
	keys := make([]uint32, len(ops))
	isDeletion := make([]uint8, (len(keys)-1)/8+1)
	for j := range keys {
		keys[j] = ops[j].Key
		if ops[j].Delete {
			isDeletion[j/8] |= (1 << (j % 8))
		}
	}
	return KeyUpdates{
		Keys:       keys,
		IsDeletion: isDeletion}
}

func (s pirServerUpdatable) returnDiffKeys(nextTimestamp int) KeyUpdates {
	firstPos := 0
	if nextTimestamp >= s.initialTimestamp {
		firstPos = nextTimestamp - s.initialTimestamp
	}
	if firstPos == len(s.ops) {
		return KeyUpdates{}
	}

	keyUpdates := opsToKeyUpdates(s.ops[firstPos:])
	keyUpdates.InitialTimestamp = int32(s.initialTimestamp + firstPos)
	return keyUpdates
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

type clientLayer struct {
	numRows int
	pir     pirClientImpl

	// debug
	hintNumBytes int
}

type pirClientUpdatable struct {
	randSource       *rand.Rand
	initialTimestamp int
	defragTimestamp  int
	numRows          int
	ops              []dbOp
	keyToPos         map[uint32]int32
	layers           []clientLayer
	hintLayers       []Layer
	servers          [2]PirServer
}

func NewPirClientUpdatable(source *rand.Rand, servers [2]PirServer) *pirClientUpdatable {
	return &pirClientUpdatable{randSource: source, servers: servers, keyToPos: make(map[uint32]int32)}
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
	hintReq := HintReq{
		Layers:          c.hintLayers,
		DefragTimestamp: int32(c.defragTimestamp),
		NextTimestamp:   int32(nextTimestamp),
		RandSeed:        int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}

	c.hintLayers = hintResp.Layers

	if hintResp.ShouldDeleteHistory {
		c.ops = []dbOp{}
		c.numRows = 0
		c.initialTimestamp = int(hintResp.KeyUpdates.InitialTimestamp)
		c.defragTimestamp = hintResp.DefragTimestamp
		c.keyToPos = make(map[uint32]int32)
	} else if hintResp.DefragTimestamp > c.defragTimestamp {
		newOps, _ := defrag(c.ops, hintResp.DefragTimestamp-c.initialTimestamp)
		c.defragTimestamp = hintResp.DefragTimestamp
		c.initialTimestamp += (len(c.ops) - len(newOps))
		c.ops = newOps
		c.numRows = 0
		c.keyToPos = make(map[uint32]int32)
		c.updatePositionMap(c.initialTimestamp)
	}
	newKeys := make([]dbOp, len(c.ops)+len(hintResp.KeyUpdates.Keys))
	copy(newKeys, c.ops)
	for i := range hintResp.KeyUpdates.Keys {
		isDelete := (hintResp.KeyUpdates.IsDeletion[i/8] & (1 << (i % 8))) != 0
		newKeys[len(c.ops)+i] = dbOp{
			Key:    hintResp.KeyUpdates.Keys[i],
			Delete: isDelete,
		}
	}
	c.ops = newKeys

	if err := c.initLayers(hintResp); err != nil {
		return err
	}

	c.updatePositionMap(int(hintResp.KeyUpdates.InitialTimestamp))
	return nil
}

func (c pirClientUpdatable) Read(key uint32) (Row, error) {
	queryReq, reconstructFunc := c.query(int(key))
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %x", key)
	}
	responses := make([]QueryResp, 2)
	err := c.servers[Left].Answer(queryReq[Left], &responses[Left])
	if err != nil {
		return nil, err
	}

	err = c.servers[Right].Answer(queryReq[Right], &responses[Right])
	if err != nil {
		return nil, err
	}
	return reconstructFunc(responses)
}

func (c *pirClientUpdatable) initLayers(resp HintResp) error {
	newLayers := make([]clientLayer, len(resp.BatchResps))
	for l, subResp := range resp.BatchResps {
		newLayers[l] = clientLayer{numRows: subResp.NumRows}
		if subResp.NumRows == 0 {
			continue
		}
		if subResp.PirType != None {
			newLayers[l].pir = NewPirClientByType(subResp.PirType, c.randSource)
			err := newLayers[l].pir.initHint(&subResp)
			if err != nil {
				return err
			}
			// Debug
			hintBytes, err := SerializedSizeOf(subResp)
			if err != nil {
				return err
			}
			newLayers[l].hintNumBytes = hintBytes
		} else {
			// Copy existing Hints for layers that haven't changed.
			newLayers[l].pir = c.layers[l].pir
			newLayers[l].hintNumBytes = c.layers[l].hintNumBytes
		}
	}

	c.layers = newLayers

	return nil
}

func (c *pirClientUpdatable) updatePositionMap(firstTimestamp int) {
	for i := firstTimestamp - c.initialTimestamp; i < len(c.ops); i++ {
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
		req[Left].BatchReqs[l].FirstRow = int32(c.hintLayers[l].FirstRow)
		req[Right].BatchReqs[l].FirstRow = int32(c.hintLayers[l].FirstRow)
		req[Left].BatchReqs[l].NumRows = int32(c.hintLayers[l].NumRows)
		req[Right].BatchReqs[l].NumRows = int32(c.hintLayers[l].NumRows)
		req[Left].BatchReqs[l].PirType = c.hintLayers[l].PirType
		req[Right].BatchReqs[l].PirType = c.hintLayers[l].PirType
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

	keyUpdates := opsToKeyUpdates(c.ops)
	keysBytes, err := SerializedSizeOf(keyUpdates)
	if err != nil {
		return 0
	}
	numBytes += keysBytes

	for _, l := range c.layers {
		numBytes += l.hintNumBytes
	}

	return numBytes
}
