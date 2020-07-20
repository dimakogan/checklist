package boosted

import (
	"fmt"
	"math/rand"
)

type partition struct {
	univSize int
	numSets  int
	setSize  int

	key []byte
	prp *PRP

	// changes
	fwd  map[int]int
	back map[int]int
}

func NewPartition(src *rand.Rand, univSize, numSets int) (*partition, error) {
	key := make([]byte, 16)
	if l, err := src.Read(key); l != len(key) || err != nil {
		panic(err)
	}
	return NewPartitionFromKey(key, univSize, numSets)
}

func NewPartitionFromKey(key []byte, univSize, numSets int) (*partition, error) {
	univSizeBits := numRecordsToUnivSizeBits(univSize)
	prp, err := NewPRP(key, univSizeBits)
	if err != nil {
		return nil, fmt.Errorf("Failed to create PRP: %s", err)
	}

	return &partition{
		prp:      prp,
		key:      key,
		univSize: univSize,
		numSets:  numSets,
		setSize:  univSize / numSets,
		fwd:      make(map[int]int),
		back:     make(map[int]int),
	}, nil
}

func (p *partition) Find(i int) int {
	return p.prp.Invert(i) / p.setSize
}

func (p *partition) Set(j int) Set {
	set := make(Set)
	for i := 0; i < p.setSize; i++ {
		set[p.prp.Eval(j*p.setSize+i)] = Present_Yes
	}
	return set
}

func (p *partition) Key() []byte {
	return p.key
}
