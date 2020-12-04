package boosted

import (
	"fmt"
	"math/rand"

	"github.com/elliotchance/orderedmap"
)

type serverLayer struct {
	// Including both deletes and adds. This is not the same as len(db).
	numOps int

	numRows int

	pir PirServer

	pirType PirType
}

type dbOp struct {
	Key    uint32
	Delete bool
	data   Row
}

type pirServerUpdatable struct {
	initialTimestamp int
	defragTimestamp  int
	ops              []dbOp
	kv               *orderedmap.OrderedMap

	layers   []serverLayer
	maxSizes []int
	pirType  PirType

	randSource *rand.Rand

	curTimestamp int32

	defragRatio int
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

func (s *pirServerUpdatable) initLayers(nRows int) {
	s.maxSizes = s.layersMaxSize(nRows)

	s.layers = make([]serverLayer, len(s.maxSizes))
	if len(s.layers) == 0 {
		return
	}

	if NumLayerActivations == nil {
		NumLayerActivations = make(map[int]int)
		NumLayerHintBytes = make(map[int]int)
	}
	for i := range s.layers {
		s.layers[i].pirType = s.pirType
		if _, ok := NumLayerActivations[i]; !ok {
			NumLayerActivations[i] = 0
		}
		if _, ok := NumLayerHintBytes[i]; !ok {
			NumLayerHintBytes[i] = 0
		}
	}

	// Even when using PirPunc, the smallest layer always uses matrix
	if s.pirType == Punc {
		s.layers[len(s.layers)-1].pirType = DPF
	}
}

func NewPirServerUpdatable(source *rand.Rand, pirType PirType) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
		defragRatio:  4,
		pirType:      pirType,
		kv:           orderedmap.NewOrderedMap(),
	}
	s.initLayers(0)
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
	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], data: rows[i]}
		s.kv.Set(keys[i], rows[i])
	}

	s.ops = append(s.ops, ops...)

	numNewRows := len(ops)
	numNewOps := len(ops)
	var i int
	for i = len(s.layers) - 1; i >= 0; i-- {
		numNewRows += s.layers[i].numRows
		numNewOps += s.layers[i].numOps
		if numNewRows <= s.maxSizes[i] {
			break
		}

		s.layers[i].numOps = 0
		s.layers[i].numRows = 0
		s.layers[i].pir = nil
	}

	if i <= 0 {
		if len(s.ops) > s.defragRatio*(s.defragTimestamp-s.initialTimestamp) {
			endDefrag := numNewOps / 2
			newOps, numRemoved := defrag(s.ops, endDefrag)
			s.defragTimestamp = s.initialTimestamp + endDefrag
			s.initialTimestamp += (len(s.ops) - len(newOps))
			s.ops = newOps
			numNewRows -= numRemoved

			numNewOps = len(s.ops)
		}
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		s.initLayers(numNewRows)
		i = 0
	}

	newRows := opsToRows(s.ops[len(s.ops)-numNewOps:])

	if len(s.layers) == 0 {
		return
	}
	s.layers[i].numOps = numNewOps
	s.layers[i].numRows = numNewRows
	s.layers[i].init(s.randSource, newRows)
}

func (s *pirServerUpdatable) DeleteRows(keys []uint32) {
	ops := make([]dbOp, len(keys))
	for i := range keys {
		ops[i] = dbOp{Key: keys[i], Delete: true}
		s.kv.Delete(keys[i])
	}
	s.ops = append(s.ops, ops...)
	s.layers[len(s.layers)-1].numOps += len(ops)
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

func (layer *serverLayer) init(randSrc *rand.Rand, db []Row) {
	layer.numRows = len(db)
	if len(db) == 0 {
		layer.pir = nil
		return
	}
	layer.pir = NewPirServerByType(layer.pirType, randSrc, db)
}

var NumLayerActivations map[int]int
var NumLayerHintBytes map[int]int

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	clientSrc := rand.New(rand.NewSource(req.RandSeed))
	resp.BatchResps = make([]HintResp, len(s.layers))
	resp.DefragTimestamp = s.defragTimestamp
	if int(req.NextTimestamp) < s.defragTimestamp {
		resp.ShouldDeleteHistory = true
		req.NextTimestamp = int32(s.initialTimestamp)
	}
	resp.KeyUpdates = s.returnDiffKeys(int(req.NextTimestamp))

	layerEnd := s.initialTimestamp
	for l, layer := range s.layers {
		layerEnd += layer.numOps
		if layer.pir != nil && (int(req.NextTimestamp) < layerEnd) {
			layer.pir.Hint(
				HintReq{RandSeed: int64(clientSrc.Uint64())},
				&resp.BatchResps[l])
			resp.BatchResps[l].PirType = layer.pirType
			NumLayerActivations[l]++
			NumLayerHintBytes[l] += len(resp.BatchResps[l].Hints) * resp.BatchResps[l].NumRowsPerBlock * resp.BatchResps[l].RowLen
		}
		resp.BatchResps[l].NumRows = layer.numRows
	}

	return nil
}

func (s pirServerUpdatable) returnDiffKeys(nextTimestamp int) KeyUpdates {
	firstPos := 0
	if nextTimestamp >= s.initialTimestamp {
		firstPos = nextTimestamp - s.initialTimestamp
	}
	if firstPos == len(s.ops) {
		return KeyUpdates{}
	}

	keys := make([]uint32, len(s.ops)-firstPos)
	isDeletion := make([]uint8, (len(keys)-1)/8+1)
	for j := range keys {
		keys[j] = s.ops[firstPos+j].Key
		if s.ops[firstPos+j].Delete {
			isDeletion[j/8] |= (1 << (j % 8))
		}
	}
	return KeyUpdates{
		InitialTimestamp: int32(s.initialTimestamp + firstPos),
		Keys:             keys,
		IsDeletion:       isDeletion}

}

func (s pirServerUpdatable) Answer(req QueryReq, resp *QueryResp) error {

	resp.BatchResps = make([]QueryResp, len(req.BatchReqs))
	for l, q := range req.BatchReqs {
		if s.layers[l].pir == nil {
			continue
		}
		// start := time.Now()
		err := s.layers[l].pir.Answer(q, &(resp.BatchResps[l]))
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
}

type pirClientUpdatable struct {
	randSource       *rand.Rand
	initialTimestamp int
	defragTimestamp  int
	numRows          int
	ops              []dbOp
	keyToPos         map[uint32]int32
	layers           []clientLayer
	servers          [2]PirServer
}

func NewPirClientUpdatable(source *rand.Rand, servers [2]PirServer) *pirClientUpdatable {
	return &pirClientUpdatable{randSource: source, servers: servers, keyToPos: make(map[uint32]int32)}
}

func (c *pirClientUpdatable) Init() error {
	return c.Update()
}

func (c *pirClientUpdatable) Update() error {
	nextTimestamp := c.nextTimestamp()
	hintReq := HintReq{
		NextTimestamp: int32(nextTimestamp),
		RandSeed:      int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}

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

func (c pirClientUpdatable) Read(i int) (Row, error) {
	queryReq, reconstructFunc := c.query(i)
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %d", i)
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
		} else {
			// Copy existing Hints for layers that haven't changed.
			newLayers[l].pir = c.layers[l].pir
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
	}
	return req, func(resps []QueryResp) (Row, error) {
		row, err := reconstructFunc([]QueryResp{
			resps[Left].BatchResps[matchingLayer],
			resps[Right].BatchResps[matchingLayer]})
		return row, err
	}
}
