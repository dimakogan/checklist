package boosted

import (
	"math/rand"
)

type clientLayer struct {
	maxSize int

	firstRow int
	numRows  int

	pirType PirType
	pir     PIRClient

	// debug
	hintNumBytes int
}

type PirClientWaterfall struct {
	pirType    PirType
	randSource *rand.Rand
	numRows    int
	rowLen     int
	layers     []clientLayer

	// For testing
	smallestLayerSizeOverride int
}

func NewPirClientWaterfall(source *rand.Rand, pirType PirType) *PirClientWaterfall {
	return &PirClientWaterfall{
		randSource: source,
		pirType:    pirType}
}

func (c *PirClientWaterfall) reset() {
	c.numRows = 0
	c.layers = nil
}

func (c *PirClientWaterfall) smallestLayerSize(nRows int) int {
	if c.smallestLayerSizeOverride != 0 {
		return c.smallestLayerSizeOverride
	}
	return 10 * SecParam * SecParam
}

func (c *PirClientWaterfall) LayersMaxSize(nRows int) []int {
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

func (c *PirClientWaterfall) freshLayers(numRows int) []clientLayer {
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

func (c *PirClientWaterfall) HintUpdateReq(numNewRows int) (*HintReq, error) {
	c.numRows += numNewRows
	layerNum := c.updateLayers(numNewRows)
	if layerNum < 0 || c.layers[layerNum].numRows == 0 {
		return nil, nil
	}
	layer := &c.layers[layerNum]
	if layer.pirType != Punc {
		layer.pir.InitHint(&HintResp{NumRows: layer.numRows, RowLen: c.rowLen})
		return nil, nil
	}
	return &HintReq{
		RandSeed: int64(c.randSource.Uint64()),
		FirstRow: layer.firstRow,
		NumRows:  layer.numRows,
		PirType:  Punc,
	}, nil
}

func (c *PirClientWaterfall) updateLayers(numNewRows int) int {
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
	if i == len(c.layers)-1 && c.pirType == Punc {
		layer.pirType = DPF
	}
	layer.pir = NewPirClientByType(layer.pirType, c.randSource)

	return i
}

func (c *PirClientWaterfall) InitHint(resp *HintResp) error {
	var l int
	for l = len(c.layers) - 1; l >= 0; l-- {
		if c.layers[l].pir != nil {
			break
		}
	}
	if l < 0 {
		panic("InitHint called with no active layers")
	}
	err := c.layers[l].pir.InitHint(resp)
	if err != nil {
		return err
	}
	// Debug
	offlineBytes, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	c.layers[l].hintNumBytes = offlineBytes
	return nil
}

func (c *PirClientWaterfall) Query(pos int) ([]QueryReq, ReconstructFunc) {
	req := make([]QueryReq, 2)
	var reconstructFunc ReconstructFunc

	var layerEnd int
	matchingLayer := 0
	for l, layer := range c.layers {
		var q []QueryReq
		if layer.pir == nil {
			continue
		}
		if layerEnd <= int(pos) && int(pos) < layerEnd+layer.numRows {
			q, reconstructFunc = layer.pir.Query(int(pos) - layerEnd)
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

func (c *PirClientWaterfall) StorageNumBytes() int {
	var numBytes int
	for _, l := range c.layers {
		numBytes += l.hintNumBytes
	}

	return numBytes
}
