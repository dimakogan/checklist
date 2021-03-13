package pir

import (
	"fmt"
	"math/rand"
)

//go:generate enumer -type=PirType
type PirType int

const (
	None PirType = iota
	Matrix
	Punc
	Perm
	DPF
	NonPrivate
)

type Server interface {
	Hint(req HintReq, resp *HintResp) error
	Answer(q QueryReq, resp *interface{}) error
}

func NewHintReq(pirType PirType) HintReq {
	switch pirType {
	case Matrix:
		return NewMatrixHintReq()
	case Punc:
		return NewPuncHintReq()
	case DPF:
		return NewDPFHintReq()
	case NonPrivate:
		return NewNonPrivateHintReq()
	}
	panic(fmt.Sprintf("Unknown PIR Type: %d", pirType))
}

type PIRReader interface {
	Init(pirType PirType) error
	Read(i int) (Row, error)
}

type pirReader struct {
	impl       Client
	servers    [2]Server
	randSource *rand.Rand
}

func NewPIRReader(source *rand.Rand, servers [2]Server) PIRReader {
	return &pirReader{servers: servers, randSource: source}
}

func (c *pirReader) Init(pirType PirType) error {
	req := NewHintReq(pirType)
	var hintResp HintResp
	if err := c.servers[Left].Hint(req, &hintResp); err != nil {
		return err
	}
	c.impl = hintResp.InitClient(c.randSource)
	return nil
}

func (c pirReader) Read(i int) (Row, error) {
	queryReq, reconstructFunc := c.impl.Query(i)
	if reconstructFunc == nil {
		return nil, fmt.Errorf("Failed to query: %d", i)
	}
	responses := make([]interface{}, 2)
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
