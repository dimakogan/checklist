package pir

import (
	"checklist/psetggm"
	"fmt"
	"math"
	"math/rand"

	"github.com/dkales/dpf-go/dpf"
)

type dpfClient struct {
	nRows int

	randSource *rand.Rand
}

func matVecProduct(db StaticDB, bitVector []byte) []byte {
	out := make(Row, db.RowLen)
	if db.RowLen == 32 {
		psetggm.XorHashesByBitVector(db.FlatDb, bitVector, out)
	} else {
		var j uint
		for j = 0; j < uint(db.NumRows); j++ {
			if ((1 << (j % 8)) & bitVector[j/8]) != 0 {
				xorInto(out, db.FlatDb[j*uint(db.RowLen):(j+1)*uint(db.RowLen)])
			}
		}
	}
	return out
}

type DPFHintReq struct {
}

func NewDPFHintReq() *DPFHintReq {
	return &DPFHintReq{}
}

func (req *DPFHintReq) Process(db StaticDB) (HintResp, error) {
	return &DPFHintResp{*db.Params()}, nil
}

type DPFHintResp struct {
	DBParams
}

func (resp *DPFHintResp) InitClient(source *rand.Rand) Client {
	return &dpfClient{randSource: source, nRows: resp.NRows}
}

type DPFQueryReq struct {
	dpf.DPFkey
}
type DPFQueryResp struct {
	Answer []byte
}

func (key *DPFQueryReq) Process(db StaticDB) (interface{}, error) {
	bitVec := dpf.EvalFull(key.DPFkey, uint64(math.Ceil(math.Log2(float64(db.NumRows)))))
	return &DPFQueryResp{matVecProduct(db, bitVec)}, nil
}

func (c *dpfClient) Query(idx int) ([]QueryReq, ReconstructFunc) {
	numBits := uint64(math.Ceil(math.Log2(float64(c.nRows))))
	qL, qR := dpf.Gen(uint64(idx), numBits)

	return []QueryReq{&DPFQueryReq{qL}, &DPFQueryReq{qR}}, func(resps []interface{}) (Row, error) {
		queryResps := make([]*DPFQueryResp, len(resps))
		var ok bool
		for i, r := range resps {
			if queryResps[i], ok = r.(*DPFQueryResp); !ok {
				return nil, fmt.Errorf("Invalid response type: %T, expected *DPFQueryResp", r)
			}
		}

		return c.reconstruct(queryResps)
	}
}

func (c *dpfClient) DummyQuery() []QueryReq {
	q, _ := c.Query(0)
	return q
}

func (c *dpfClient) StateSize() (int, int) {
	return 0, 0
}

func (c *dpfClient) reconstruct(resp []*DPFQueryResp) (Row, error) {
	out := make([]byte, len(resp[Left].Answer))
	xorInto(out, resp[Left].Answer)
	xorInto(out, resp[Right].Answer)
	return out, nil
}
