package pir

import (
	"fmt"
	"math"
	"math/rand"
)

type matrixClient struct {
	nRows  int
	height int
	width  int
	rowLen int

	randSource *rand.Rand
}

func getHeightWidth(nRows int, rowLen int) (int, int) {
	// h^2 = n * rowlen
	width := int(math.Ceil(math.Sqrt(float64(nRows*rowLen)) / float64(rowLen)))
	height := (nRows-1)/width + 1

	return width, height
}

func matBoolVecProduct(db StaticDB, bitVector []bool) []byte {
	width, height := getHeightWidth(db.NumRows, db.RowLen)
	out := make([]byte, width*db.RowLen)

	cnt := 0
	tableWidth := db.RowLen * width
	flatDb := db.Slice(0, db.NumRows)
	for j := 0; j < height; j++ {
		if bitVector[j] {
			start := tableWidth * j
			length := tableWidth
			if start+length >= len(flatDb) {
				length = len(flatDb) - start
			}
			xorInto(out[0:length], flatDb[start:start+length])
			cnt = cnt + tableWidth
		}
	}
	return out
}

type MatrixQueryReq struct {
	BitVector []bool
}

type MatrixQueryResp struct {
	Answer []byte
}

func (req *MatrixQueryReq) Process(db StaticDB) (interface{}, error) {
	return &MatrixQueryResp{matBoolVecProduct(db, req.BitVector)}, nil
}

type MatrixHintReq struct{}
type MatrixHintResp struct {
	DBParams
}

func NewMatrixHintReq() *MatrixHintReq {
	return &MatrixHintReq{}
}

func (req *MatrixHintReq) Process(db StaticDB) (HintResp, error) {
	return &MatrixHintResp{*db.Params()}, nil
}

func (resp *MatrixHintResp) InitClient(source *rand.Rand) Client {
	client := matrixClient{
		randSource: source,
		nRows:      resp.NRows,
		rowLen:     resp.RowLen}
	client.width, client.height = getHeightWidth(resp.NRows, client.rowLen)
	return &client
}

func (c *matrixClient) Query(idx int) ([]QueryReq, ReconstructFunc) {
	rowNum := idx / c.width
	colNum := idx % c.width
	qL := make([]bool, c.height)
	qR := make([]bool, c.height)
	for i := 0; i < c.height; i++ {
		qL[i] = (c.randSource.Uint64()&1 == 0)
		qR[i] = (qL[i] != (i == rowNum))
	}

	return []QueryReq{&MatrixQueryReq{qL}, &MatrixQueryReq{qR}}, func(resps []interface{}) (Row, error) {
		queryResps := make([]*MatrixQueryResp, len(resps))
		var ok bool
		for i, r := range resps {
			if queryResps[i], ok = r.(*MatrixQueryResp); !ok {
				return nil, fmt.Errorf("Invalid response type: %T, expected: MatrixQueryResp", r)
			}
		}
		return c.reconstruct(colNum, queryResps)
	}
}

func (c *matrixClient) DummyQuery() []QueryReq {
	q, _ := c.Query(0)
	return q
}

func (c *matrixClient) reconstruct(colNum int, resp []*MatrixQueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out[c.rowLen*colNum : (c.rowLen * (colNum + 1))], nil
}

func (c *matrixClient) StateSize() (int, int) {
	return 0, 0
}
