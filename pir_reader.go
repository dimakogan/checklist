package boosted

import (
	"fmt"
	"math/rand"
)

type PIRReader interface {
	Init(pirType PirType) error
	Read(i int) (Row, error)
}

type pirReader struct {
	impl       PIRClient
	servers    [2]PirServer
	randSource *rand.Rand
}

func NewPIRReader(source *rand.Rand, servers [2]PirServer) PIRReader {
	return &pirReader{servers: servers, randSource: source}
}

func (c *pirReader) Init(pirType PirType) error {
	c.impl = NewPirClientByType(pirType, c.randSource)
	hintReq := HintReq{RandSeed: int64(c.randSource.Uint64())}
	var hintResp HintResp
	if err := c.servers[Left].Hint(hintReq, &hintResp); err != nil {
		return err
	}
	return c.impl.InitHint(&hintResp)
}

func (c pirReader) Read(i int) (Row, error) {
	queryReq, reconstructFunc := c.impl.Query(i)
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
