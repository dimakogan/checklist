package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type serverLayer struct {
	maxSize   int
	timedRows []TimedRow
	pir       PirServer
}

type pirServerUpdatable struct {
	layers     []serverLayer
	randSource *rand.Rand

	curTimestamp int
}

func layersMaxSize(nRows int) []int {
	numLayers := int(math.Ceil(math.Log2(float64(nRows)))) + 1
	maxSize := make([]int, numLayers)
	for l := range maxSize {
		maxSize[l] = (1 << l)
	}
	return maxSize
}

func rowsToRawDb(timedRows []TimedRow) []Row {
	processedRows := processDeletes(timedRows)
	db := make([]Row, len(processedRows))
	for i, timedRow := range processedRows {
		db[i] = timedRow.data
	}
	return db
}

func processDeletes(timedRows []TimedRow) []TimedRow {
	isDeleted := make([]bool, len(timedRows))
	keyToPos := make(map[uint32]int, len(timedRows))
	for i, row := range timedRows {
		if row.Delete {
			// Delete the 'delete' row itself
			isDeleted[i] = true
			// Check for an unexisting row.
			// This can happen since a delete may be in a different layer
			// than an earlier add.
			if _, ok := keyToPos[row.Key]; ok {
				isDeleted[keyToPos[row.Key]] = true
			}
		} else {
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
	}
	return processedRows
}

func (s *pirServerUpdatable) initLayers(timedRows []TimedRow) {
	db := rowsToRawDb(timedRows)
	maxSizes := layersMaxSize(len(db))

	s.layers = make([]serverLayer, len(maxSizes))
	for l := range s.layers {
		s.layers[l] = serverLayer{
			pir:       nil,
			timedRows: []TimedRow{},
			maxSize:   maxSizes[l],
		}
	}
	// Initially, store all data in last (biggest) layer
	s.layers[len(maxSizes)-1].timedRows = timedRows
	s.layers[len(maxSizes)-1].pir = NewPirServerPunc(s.randSource, db)
}

func NewPirServerUpdatable(source *rand.Rand, keys []uint32, values []Row) *pirServerUpdatable {
	s := pirServerUpdatable{
		randSource:   source,
		curTimestamp: 0,
	}
	timedRows := make([]TimedRow, 0, len(keys))
	for i, key := range keys {
		timedRows = append(timedRows,
			TimedRow{Timestamp: s.tick(), Key: key, data: values[i]})
	}
	s.initLayers(timedRows)
	return &s
}

func (s *pirServerUpdatable) tick() int {
	s.curTimestamp++
	return s.curTimestamp
}

func (s *pirServerUpdatable) AddRow(key uint32, row Row) {
	newRows := []TimedRow{
		{Timestamp: s.tick(), Key: key, data: row},
	}
	var i int
	for i = 0; i < len(s.layers); i++ {
		newRows = append(s.layers[i].timedRows, newRows...)
		s.layers[i].timedRows = []TimedRow{}
		s.layers[i].pir = nil
		if len(processDeletes(newRows)) <= s.layers[i].maxSize {
			break
		}

	}
	if i >= len(s.layers) {
		// Biggest layer reached capacity.
		// Recompute all layer sizes and reinitialize
		s.initLayers(newRows)
	} else {
		s.layers[i].timedRows = newRows
		s.layers[i].pir = NewPirServerPunc(s.randSource, rowsToRawDb(newRows))
	}
}

func (layer serverLayer) newLayerHint(req HintReq, resp *HintResp) bool {
	earliestNewKey := len(layer.timedRows)
	for {
		if earliestNewKey == 0 {
			break
		}
		if layer.timedRows[earliestNewKey-1].Timestamp <= req.LatestKeyTimestamp {
			break
		}
		earliestNewKey--
	}
	resp.TimedKeys = make([]TimedRow, len(layer.timedRows)-earliestNewKey)
	if len(resp.TimedKeys) == 0 {
		return false
	}
	for j := 0; j < len(resp.TimedKeys); j++ {
		resp.TimedKeys[j] = layer.timedRows[earliestNewKey+j]
		resp.TimedKeys[j].data = nil
	}
	if layer.pir != nil {
		layer.pir.Hint(req, resp)
	}
	return true
}

func (s pirServerUpdatable) Hint(req HintReq, resp *HintResp) error {
	clientSrc := rand.New(rand.NewSource(req.RandSeed))
	resp.BatchResps = make([]HintResp, len(s.layers))
	unchanged := true
	resp.NumUnchangedLayers = 0
	for l := len(s.layers) - 1; l >= 0; l-- {
		isNew := s.layers[l].newLayerHint(HintReq{LatestKeyTimestamp: req.LatestKeyTimestamp, RandSeed: int64(clientSrc.Uint64())},
			&resp.BatchResps[l])
		unchanged = unchanged && !isNew
		if unchanged {
			resp.NumUnchangedLayers++
		}

	}
	resp.BatchResps = resp.BatchResps[0 : len(resp.BatchResps)-resp.NumUnchangedLayers]
	return nil
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
	timedKeys []TimedRow
	pir       pirClientImpl
}

type rowLayerPosition struct {
	layer, posInLayer int
}

type pirClientUpdatable struct {
	randSource *rand.Rand
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
	hintReq := HintReq{
		LatestKeyTimestamp: c.latestKeyTimestamp(),
		RandSeed:           int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}
	if err := c.initLayers(hintResp); err != nil {
		return err
	}
	c.updatePositionMap()
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
	newLayers := make([]clientLayer, len(resp.BatchResps)+resp.NumUnchangedLayers)
	for l, subResp := range resp.BatchResps {
		newLayers[l] = clientLayer{timedKeys: subResp.TimedKeys}
		if subResp.NumRows == 0 {
			continue
		}
		newLayers[l].pir = NewPirClientPunc(c.randSource)
		err := newLayers[l].pir.initHint(&subResp)
		if err != nil {
			return err
		}
	}
	existingRows := []TimedRow{}
	for l := 0; l < len(c.layers)-resp.NumUnchangedLayers; l++ {
		existingRows = append(c.layers[l].timedKeys, existingRows...)
	}

	if resp.NumUnchangedLayers < len(c.layers) {
		newLayers[len(newLayers)-resp.NumUnchangedLayers-1].timedKeys =
			append(existingRows, newLayers[len(newLayers)-resp.NumUnchangedLayers-1].timedKeys...)
	}
	copy(newLayers[len(newLayers)-resp.NumUnchangedLayers:],
		c.layers[len(c.layers)-resp.NumUnchangedLayers:])
	c.layers = newLayers
	return nil
}

func (c *pirClientUpdatable) updatePositionMap() {
	c.positions = make(map[uint32]rowLayerPosition)
	for l, layer := range c.layers {
		processedKeys := processDeletes(layer.timedKeys)
		for i, elem := range processedKeys {
			c.positions[elem.Key] = rowLayerPosition{l, i}
		}
	}
}

func (c *pirClientUpdatable) latestKeyTimestamp() int {
	for i := 0; i < len(c.layers); i++ {
		layerKeys := c.layers[i].timedKeys
		if len(layerKeys) > 0 {
			return layerKeys[len(layerKeys)-1].Timestamp
		}
	}
	return -1
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
