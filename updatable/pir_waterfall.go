package updatable

import (
	"fmt"
	"math/rand"

	"checklist/pir"
)

type clientLayer struct {
	maxSize int

	firstRow int
	numRows  int

	pirType pir.PirType
	pir     pir.Client
}

type WaterfallClient struct {
	pirType    pir.PirType
	randSource *rand.Rand
	numRows    int
	rowLen     int
	layers     []clientLayer

	// For testing
	smallestLayerSizeOverride int
}

func NewWaterfallClient(source *rand.Rand, pirType pir.PirType) *WaterfallClient {
	return &WaterfallClient{
		randSource: source,
		pirType:    pirType}
}

func (c *WaterfallClient) reset() {
	c.numRows = 0
	c.layers = nil
}

func (c *WaterfallClient) smallestLayerSize(nRows int) int {
	if c.smallestLayerSizeOverride != 0 {
		return c.smallestLayerSizeOverride
	}
	return 10 * pir.SecParam * pir.SecParam
}

func (c *WaterfallClient) LayersMaxSize(nRows int) []int {
	// if nRows == 0 {
	// 	return []int{}
	// }
	if c.pirType != pir.Punc {
		return []int{nRows}
	}
	maxSize := []int{nRows}
	smallest := c.smallestLayerSize(nRows)
	for maxSize[len(maxSize)-1] > smallest {
		maxSize = append(maxSize, maxSize[len(maxSize)-1]/2)
	}
	return maxSize
}

func (c *WaterfallClient) freshLayers(numRows int) []clientLayer {
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

type UpdatableHintReq struct {
	Req      pir.HintReq
	FirstRow int
	NumRows  int
}

func (req *UpdatableHintReq) Process(db pir.StaticDB) (pir.HintResp, error) {
	layerFlatDb := db.Slice(req.FirstRow, req.FirstRow+req.NumRows)
	return req.Req.Process(pir.StaticDB{NumRows: req.NumRows, RowLen: db.RowLen, FlatDb: layerFlatDb})
}

func (c *WaterfallClient) HintUpdateReq(numNewRows int, rowLen int) (*UpdatableHintReq, error) {
	c.numRows += numNewRows
	c.rowLen = rowLen
	layerNum := c.updateLayers(numNewRows)
	if layerNum < 0 || c.layers[layerNum].numRows == 0 {
		return nil, nil
	}
	layer := &c.layers[layerNum]
	if layer.pirType != pir.Punc {
		req, err := pir.NewHintReq(c.randSource, layer.pirType).Process(pir.StaticDB{NumRows: layer.numRows, RowLen: c.rowLen})
		if err != nil {
			return nil, err
		}
		layer.pir = req.InitClient(c.randSource)
		return nil, nil
	}
	return &UpdatableHintReq{
		FirstRow: layer.firstRow,
		NumRows:  layer.numRows,
		Req:      pir.NewPuncHintReq(c.randSource),
	}, nil
}

func (c *WaterfallClient) updateLayers(numNewRows int) int {
	var i int
	for i = len(c.layers) - 1; i >= 0; i-- {
		numNewRows += c.layers[i].numRows
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
	layer.numRows = numNewRows
	layer.firstRow = c.numRows - numNewRows
	layer.pirType = c.pirType
	if i == len(c.layers)-1 && c.pirType == pir.Punc {
		layer.pirType = pir.DPF
	}
	layer.pir = nil

	return i
}

func (c *WaterfallClient) InitHint(resp pir.HintResp) error {
	var l int
	for l = len(c.layers) - 1; l >= 0; l-- {
		if c.layers[l].numRows != 0 {
			break
		}
	}
	if l < 0 {
		panic("InitHint called with no active layers")
	}
	c.layers[l].pir = resp.InitClient(c.randSource)
	return nil
}

type UpdatableQueryReq struct {
	Reqs     []pir.QueryReq
	FirstRow []int
}

type UpdatableQueryResp []interface{}

func (req UpdatableQueryReq) Process(db pir.StaticDB) (interface{}, error) {
	resps := make([]interface{}, len(req.Reqs))
	var err error
	for l, q := range req.Reqs {
		layerFlatDb := db.Slice(req.FirstRow[l], req.FirstRow[l+1])
		resps[l], err = q.Process(pir.StaticDB{req.FirstRow[l+1] - req.FirstRow[l], db.RowLen, layerFlatDb})
		if err != nil {
			return nil, err
		}
	}
	return resps, nil
}

func (c *WaterfallClient) Query(pos int) ([]pir.QueryReq, pir.ReconstructFunc) {
	req := make([]UpdatableQueryReq, 2)
	var reconstructFunc pir.ReconstructFunc

	var layerEnd int
	matchingLayer := 0
	for l, layer := range c.layers {
		var q []pir.QueryReq
		if layer.pir == nil {
			continue
		}
		if layerEnd <= int(pos) && int(pos) < layerEnd+layer.numRows {
			q, reconstructFunc = layer.pir.Query(int(pos) - layerEnd)
			matchingLayer = len(req[pir.Left].Reqs)
		} else {
			q = layer.pir.DummyQuery()
		}
		for s := range []int{pir.Left, pir.Right} {
			req[s].FirstRow = append(req[s].FirstRow, c.layers[l].firstRow)
			req[s].Reqs = append(req[s].Reqs, q[s])
		}
		layerEnd += layer.numRows
	}
	for s := range []int{pir.Left, pir.Right} {
		req[s].FirstRow = append(req[s].FirstRow, layerEnd)
	}
	return []pir.QueryReq{req[pir.Left], req[pir.Right]}, func(resps []interface{}) (pir.Row, error) {
		queryResps := make([][]interface{}, len(resps))
		var ok bool
		for i, r := range resps {
			if queryResps[i], ok = r.([]interface{}); !ok {
				return nil, fmt.Errorf("Invalid response type: %T, expected []interface{}", r)
			}
		}

		row, err := reconstructFunc([]interface{}{
			queryResps[pir.Left][matchingLayer],
			queryResps[pir.Right][matchingLayer]})
		return row, err
	}
}

func (c *WaterfallClient) State() (bitsPerKey, fixedBytes int) {
	for _, layer := range c.layers {
		if layer.pir != nil {
			bpk, fixed := layer.pir.StateSize()
			if bpk > bitsPerKey {
				bitsPerKey = bpk
			}
			fixedBytes += fixed
		}
	}
	return
}
