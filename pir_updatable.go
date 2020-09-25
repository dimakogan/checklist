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

	pir PirServer

	isMatrix bool
}

type pirServerUpdatable struct {
	timedRows []TimedRow

	layers    []serverLayer
	maxSizes  []int
	useMatrix bool

	randSource *rand.Rand

	curTimestamp    int
	defragTimestamp int

	defragRatio int
}

func (s pirServerUpdatable) layersMaxSize(nRows int) []int {
	if nRows == 0 {
		return []int{}
	}
	maxSize := []int{int(math.Round(math.Sqrt(float64(nRows))))}
	for maxSize[len(maxSize)-1] < nRows {
		maxSize = append(maxSize, 2*maxSize[len(maxSize)-1])
	}
	return maxSize
}

func rowsToRawDb(rows []TimedRow) []Row {
	db := make([]Row, len(rows))
	for i, timedRow := range rows {
		db[i] = timedRow.data
	}
	return db
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
		s.layers[i].isMatrix = s.useMatrix
	}

	// The smallest layer always uses matrix
	s.layers[len(s.layers)-1].isMatrix = true
}

func NewPirServerUpdatable(source *rand.Rand, useMatrix bool) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
		defragRatio:  4,
		useMatrix:    useMatrix,
	}
	return &s
}

func (s *pirServerUpdatable) tick() int {
	s.curTimestamp++
	return s.curTimestamp
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
	s.timedRows = append(s.timedRows, timedRows...)
	numNewRows := len(timedRows)
	var i int
	for i = len(s.layers) - 1; i >= 0; i-- {
		numNewRows += s.layers[i].numTimedRows
		if numNewRows <= s.maxSizes[i] {
			break
		}

		s.layers[i].numTimedRows = 0
		s.layers[i].pir = nil
	}
	processedRows, _ := processDeletes(s.timedRows[len(s.timedRows)-numNewRows:])

	if i <= 0 {
		// If the the number of deletions ovewhelms the actual size of the DB, then
		// `defrag` the database.
		if s.defragRatio*len(processedRows) < numNewRows {
			s.defrag(len(processedRows) * s.defragRatio / 2)
			numNewRows = len(s.timedRows)
		}
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		s.initLayers(len(processedRows))
		i = 0
	}

	if len(s.layers) == 0 {
		return
	}
	s.layers[i].numTimedRows = numNewRows
	s.layers[i].init(s.randSource, rowsToRawDb(processedRows))
}

func (s *pirServerUpdatable) defrag(numRowsToFree int) {
	defragRows := make(map[uint32]int)
	var i int
	var row TimedRow
	for i, row = range s.timedRows {
		if row.Delete {
			delete(defragRows, row.Key)
		} else {
			defragRows[row.Key] = i
		}
		if i-len(defragRows)+1 >= numRowsToFree {
			break
		}
	}
	s.defragTimestamp = s.timedRows[i].Timestamp
	// Push forward all defragged rows
	pos := i
	newTs := s.defragTimestamp
	for ; i >= 0; i-- {
		if i2, ok := defragRows[s.timedRows[i].Key]; ok && i2 == i {
			s.timedRows[pos] = s.timedRows[i]
			s.timedRows[pos].Timestamp = newTs
			newTs--
			pos--
		}
	}
	s.timedRows = s.timedRows[pos+1:]
}

func (layer *serverLayer) init(randSrc *rand.Rand, db []Row) {
	if len(db) == 0 {
		layer.pir = nil
		return
	}
	if layer.isMatrix {
		layer.pir = NewPirServerMatrix(randSrc, db)
	} else {
		layer.pir = NewPirServerPunc(randSrc, db)
	}
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
			resp.BatchResps[l].IsMatrix = layer.isMatrix
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
		err := s.layers[l].pir.Answer(q, &(resp.BatchResps[l]))
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
		if subResp.IsMatrix {
			newLayers[l].pir = NewPirClientMatrix(c.randSource)
		} else {
			newLayers[l].pir = NewPirClientPunc(c.randSource)
		}
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
	reconstructFuncs := make([]ReconstructFunc, len(c.layers))

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
		q, reconstructFuncs[l] = layer.pir.query(iPos.posInLayer)
		req[Left].BatchReqs[l] = q[Left]
		req[Right].BatchReqs[l] = q[Right]
	}
	return req, func(resps []QueryResp) (Row, error) {
		var ans Row
		for l, f := range reconstructFuncs {
			if f == nil {
				continue
			}
			row, err := f([]QueryResp{
				resps[Left].BatchResps[l],
				resps[Right].BatchResps[l]})
			if err != nil {
				return row, err
			}
			if l == iPos.layer {
				ans = row
			}
		}
		return ans, nil
	}
}
