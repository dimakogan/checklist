package boosted

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type serverLayer struct {
	// Including both deletes and adds. This is not the same as len(db).
	numTimedRows int

	numRows int

	pir PirServer

	pirType PirType
}

type pirServerUpdatable struct {
	timedRows []TimedRow
	keyToPos  map[uint32]int

	layers   []serverLayer
	maxSizes []int
	pirType  PirType

	randSource *rand.Rand

	curTimestamp    int
	defragTimestamp int

	defragRatio int
}

func smallestLayerSize(nRows int) int {
	return int(math.Round(math.Sqrt(float64(nRows))))
}

func (s pirServerUpdatable) layersMaxSize(nRows int) []int {
	if nRows == 0 {
		return []int{}
	}
	if s.pirType == Matrix {
		return []int{nRows}
	}
	maxSize := []int{nRows}
	smallest := smallestLayerSize(nRows)
	for maxSize[len(maxSize)-1] > smallest {
		maxSize = append(maxSize, maxSize[len(maxSize)-1]/4)
	}
	return maxSize
}

func rowsToRawDb(rows []TimedRow) []Row {
	db := make([]Row, 0, len(rows))
	for _, timedRow := range rows {
		if timedRow.Delete || timedRow.DeletedTimestamp > 0 {
			continue
		}
		db = append(db, timedRow.data)
	}
	return db
}

func (s *pirServerUpdatable) processUpdates(timedRows []TimedRow) {
	prevLen := len(s.timedRows)
	s.timedRows = append(s.timedRows, timedRows...)
	for i, row := range timedRows {
		if pos, ok := s.keyToPos[row.Key]; ok {
			s.timedRows[pos].DeletedTimestamp = row.Timestamp
			delete(s.keyToPos, row.Key)
		}
		if !row.Delete {
			s.keyToPos[row.Key] = prevLen + i
		}
	}
}

func processDeletes(timedRows []TimedRow) (adds, deletes []TimedRow) {
	isDeleted := make([]bool, len(timedRows))
	keyToPos := make(map[uint32]int, len(timedRows))
	for i, row := range timedRows {
		if row.Delete {
			// Delete the 'delete' row itself
			isDeleted[i] = true
			// Check for an unexisting row.
			// This can happen since a delete may be in a different layer
			// than an earlier add.
			if pos, ok := keyToPos[row.Key]; ok {
				isDeleted[pos] = true
				delete(keyToPos, row.Key)
			}
		} else {
			if prevPos, ok := keyToPos[row.Key]; ok {
				isDeleted[prevPos] = true
			}
			keyToPos[row.Key] = i
		}
	}
	length := 0

	for i := range isDeleted {
		if !isDeleted[i] {
			length++
		}
	}
	processedRows := make([]TimedRow, 0, length)
	for i, row := range timedRows {
		if !isDeleted[i] {
			processedRows = append(processedRows, row)
		}
		if row.Delete {
			deletes = append(deletes, row)
		}
	}
	return processedRows, deletes
}

func (s *pirServerUpdatable) initLayers(nRows int) {
	s.maxSizes = s.layersMaxSize(nRows)

	s.layers = make([]serverLayer, len(s.maxSizes))
	if len(s.layers) == 0 {
		return
	}

	for i := range s.layers {
		s.layers[i].pirType = s.pirType
	}

	// The smallest layer always uses matrix
	s.layers[len(s.layers)-1].pirType = Matrix
}

func NewPirServerUpdatable(source *rand.Rand, pirType PirType) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
		defragRatio:  4,
		pirType:      pirType,
		keyToPos:     make(map[uint32]int),
	}
	return &s
}

func (s *pirServerUpdatable) tick() int {
	s.curTimestamp++
	return s.curTimestamp
}

func (s *pirServerUpdatable) GetRow(idx int, row *RowIndexVal) error {
	keys, rows := s.elements(idx, idx+1)
	if keys == nil {
		return fmt.Errorf("Index %d out of bounds", idx)
	}
	if len(keys) != 1 || len(rows) != 1 {
		panic(fmt.Sprintf("Invalid returned slice length: %d, %d", len(keys), len(rows)))
	}

	row.Key = keys[0]
	row.Value = rows[0]
	row.Index = idx

	return nil
}

func (s *pirServerUpdatable) elements(start, end int) (keys []uint32, rows []Row) {
	if end == -1 {
		keys = make([]uint32, 0)
		rows = make([]Row, 0)
	} else {
		keys = make([]uint32, 0, end-start)
		rows = make([]Row, 0, end-start)
	}
	pos := 0
	for _, row := range s.timedRows {
		if pos >= end {
			break
		}
		if row.Delete || row.DeletedTimestamp > 0 {
			continue
		}

		if pos >= start {
			keys = append(keys, row.Key)
			rows = append(rows, row.data)
		}
		pos++
	}
	if pos < end {
		return nil, nil
	}
	return keys, rows
}

func (s *pirServerUpdatable) SomeKeys(num int) []uint32 {
	keys := make([]uint32, 0, num)
	for key := range s.keyToPos {
		keys = append(keys, key)
		if len(keys) == num {
			return keys
		}
	}
	return keys
}

func (s *pirServerUpdatable) AddRows(keys []uint32, rows []Row) {
	timedRows := make([]TimedRow, len(keys))
	for i := range keys {
		timedRows[i] = TimedRow{Timestamp: s.tick(), Key: keys[i], data: rows[i]}
	}

	s.updateLayers(timedRows)
}

func (s *pirServerUpdatable) DeleteRows(keys []uint32) {
	timedRows := make([]TimedRow, len(keys))
	for i := range keys {
		timedRows[i] = TimedRow{Timestamp: s.tick(), Key: keys[i], Delete: true}
	}
	s.updateLayers(timedRows)
}

func (s *pirServerUpdatable) updateLayers(timedRows []TimedRow) {
	s.processUpdates(timedRows)
	numNewRows := len(timedRows)
	var i int
	for i = len(s.layers) - 1; i >= 0; i-- {
		numNewRows += s.layers[i].numTimedRows
		if numNewRows <= s.maxSizes[i] {
			break
		}

		s.layers[i].numTimedRows = 0
		s.layers[i].numRows = 0
		s.layers[i].pir = nil
	}

	rawDB := rowsToRawDb(s.timedRows[len(s.timedRows)-numNewRows:])

	if i <= 0 {
		// If the the number of deletions ovewhelms the actual size of the DB, then
		// `defrag` the database.
		if s.defragRatio*len(rawDB) < numNewRows {
			s.defrag(len(rawDB) * s.defragRatio / 2)
			numNewRows = len(s.timedRows)
		}
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		s.initLayers(len(rawDB))
		i = 0
	}

	if len(s.layers) == 0 {
		return
	}
	s.layers[i].numTimedRows = numNewRows
	s.layers[i].init(s.randSource, rawDB)
}

func (s *pirServerUpdatable) defrag(numRowsToFree int) {
	freed := 0
	endDefrag := 0
	for i, row := range s.timedRows {
		if !row.Delete && row.DeletedTimestamp <= 0 {
			s.keyToPos[row.Key] -= freed
		} else if freed < numRowsToFree {
			freed++
			endDefrag = i
		}
	}
	s.defragTimestamp = s.timedRows[endDefrag].Timestamp
	// Push forward all defragged rows
	pos := endDefrag
	newTs := s.defragTimestamp
	for i := endDefrag; i >= 0; i-- {
		if !s.timedRows[i].Delete && s.timedRows[i].DeletedTimestamp <= 0 {
			s.timedRows[pos] = s.timedRows[i]
			s.timedRows[pos].Timestamp = newTs
			newTs--
			pos--
		}
	}
	s.timedRows = s.timedRows[pos+1:]
}

func (layer *serverLayer) init(randSrc *rand.Rand, db []Row) {
	layer.numRows = len(db)
	if len(db) == 0 {
		layer.pir = nil
		return
	}
	layer.pir = NewPirServerByType(layer.pirType, randSrc, db)
}

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	clientSrc := rand.New(rand.NewSource(req.RandSeed))
	resp.BatchResps = make([]HintResp, len(s.layers))
	resp.ShouldDeleteHistory = (req.LatestKeyTimestamp < s.defragTimestamp)
	resp.TimedKeys = s.returnDiffKeys(req.LatestKeyTimestamp)

	layerEnd := 0
	for l, layer := range s.layers {
		layerEnd += layer.numTimedRows
		if layer.pir != nil && s.timedRows[layerEnd-1].Timestamp > req.LatestKeyTimestamp {
			layer.pir.Hint(
				HintReq{RandSeed: int64(clientSrc.Uint64())},
				&resp.BatchResps[l])
			resp.BatchResps[l].PirType = layer.pirType
		}
		resp.BatchResps[l].EndTimestamp = s.timedRows[layerEnd-1].Timestamp
	}

	return nil
}

func (s pirServerUpdatable) returnDiffKeys(latestTimestamp int) []TimedRow {
	if latestTimestamp < s.defragTimestamp {
		return s.timedRows
	}
	earliestNewKey := len(s.timedRows)
	for {
		if earliestNewKey == 0 {
			break
		}
		if s.timedRows[earliestNewKey-1].Timestamp <= latestTimestamp {
			break
		}
		earliestNewKey--
	}
	diffKeys := make([]TimedRow, len(s.timedRows)-earliestNewKey)
	for j := 0; j < len(diffKeys); j++ {
		diffKeys[j] = s.timedRows[earliestNewKey+j]
		diffKeys[j].data = nil
	}
	return diffKeys
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
	endTimestamp int
	pir          pirClientImpl
}

type rowLayerPosition struct {
	layer, posInLayer int
}

type pirClientUpdatable struct {
	randSource *rand.Rand
	timedKeys  []TimedRow
	positions  map[uint32]rowLayerPosition
	layers     []clientLayer
	servers    [2]PirServer
}

func NewPirClientUpdatable(source *rand.Rand, servers [2]PirServer) *pirClientUpdatable {
	return &pirClientUpdatable{randSource: source, servers: servers}
}

func (c *pirClientUpdatable) Init() error {
	return c.Update()
}

func (c *pirClientUpdatable) Update() error {
	latestKeyTimestamp := c.latestKeyTimestamp()
	hintReq := HintReq{
		LatestKeyTimestamp: latestKeyTimestamp,
		RandSeed:           int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}

	if hintResp.ShouldDeleteHistory {
		c.timedKeys = []TimedRow{}
		c.positions = make(map[uint32]rowLayerPosition)
	}
	c.timedKeys = append(c.timedKeys, hintResp.TimedKeys...)

	if err := c.initLayers(hintResp); err != nil {
		return err
	}

	c.recomputePositionMap(latestKeyTimestamp)
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
		newLayers[l] = clientLayer{endTimestamp: subResp.EndTimestamp}
		if subResp.NumRows == 0 {
			continue
		}

		newLayers[l].pir = NewPirClientByType(subResp.PirType, c.randSource)
		err := newLayers[l].pir.initHint(&subResp)
		if err != nil {
			return err
		}
	}

	// Copy existing Hints for layers that haven't changed.
	// This can only happen if the number of layers is unchanged,
	// since otherwise all layers have been recomputed.
	if len(newLayers) == len(c.layers) {
		for l := range c.layers {
			if newLayers[l].endTimestamp == c.layers[l].endTimestamp {
				newLayers[l].pir = c.layers[l].pir
			}
		}
	}

	c.layers = newLayers

	return nil
}

func (c *pirClientUpdatable) recomputePositionMap(latestKeyTimestamp int) {
	layerEnd := 0
	for l := range c.layers {
		layerStart := layerEnd
		layerEnd = sort.Search(len(c.timedKeys), func(i int) bool {
			return c.layers[l].endTimestamp < c.timedKeys[i].Timestamp
		})
		// If a layer has not changed relative to previous update, no need to recompute.
		if c.layers[l].endTimestamp < latestKeyTimestamp {
			continue
		}
		processedKeys, deletedRows := processDeletes(c.timedKeys[layerStart:layerEnd])
		// The first (oldest) layer can always be defragmented on the client end
		if l == 0 {
			c.timedKeys = append(processedKeys, c.timedKeys[layerEnd:]...)
			layerEnd = len(processedKeys)
		}

		for _, row := range deletedRows {
			// propagate deletes backwards to previous layers
			delete(c.positions, row.Key)
		}
		for i, row := range processedKeys {
			c.positions[row.Key] = rowLayerPosition{l, i}
		}
	}
}

func (c *pirClientUpdatable) latestKeyTimestamp() int {
	if len(c.timedKeys) <= 0 {
		return -1
	}
	return c.timedKeys[len(c.timedKeys)-1].Timestamp
}

func (c *pirClientUpdatable) query(i int) ([]QueryReq, ReconstructFunc) {
	req := []QueryReq{
		{
			LatestKeyTimestamp: c.latestKeyTimestamp(),
			BatchReqs:          make([]QueryReq, len(c.layers))},
		{
			LatestKeyTimestamp: c.latestKeyTimestamp(),
			BatchReqs:          make([]QueryReq, len(c.layers))},
	}
	var reconstructFunc ReconstructFunc

	// Slight hack for now: using i as key
	iPos, ok := c.positions[uint32(i)]
	if !ok {
		return nil, nil
	}
	for l, layer := range c.layers {
		var q []QueryReq
		if layer.pir == nil {
			continue
		}
		if l == iPos.layer {
			q, reconstructFunc = layer.pir.query(iPos.posInLayer)
		} else {
			q = layer.pir.dummyQuery()
		}

		req[Left].BatchReqs[l] = q[Left]
		req[Right].BatchReqs[l] = q[Right]
	}
	return req, func(resps []QueryResp) (Row, error) {
		row, err := reconstructFunc([]QueryResp{
			resps[Left].BatchResps[iPos.layer],
			resps[Right].BatchResps[iPos.layer]})
		return row, err
	}
}
