package boosted

import (
	"fmt"
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

func NewPartition(key []byte, univSize, numSets int) (*partition, error) {
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

func (p *partition) Find(i int) (setNum int, posInSet int) {
	pos := p.findPos(i)
	return (pos / p.setSize), (pos % p.setSize)
}

func (p *partition) Set(j int) Set {
	set := make(Set, p.setSize)
	for i := 0; i < p.setSize; i++ {
		set[i] = p.elemAt(j*p.setSize + i)
	}
	return set
}

func (p *partition) Key() []byte {
	return p.key
}

func (p *partition) elemAt(pos int) int {
	if i, ok := p.fwd[pos]; ok {
		return i
	} else {
		return p.prp.Eval(pos)
	}
}

func (p *partition) findPos(i int) int {
	if pos, ok := p.back[i]; ok {
		return pos
	} else {
		return p.prp.Invert(i)
	}
}

func (p *partition) Swap(i, j int) {
	iPos := p.findPos(i)
	jPos := p.findPos(j)
	p.fwd[iPos] = j
	p.fwd[jPos] = i
	p.back[i] = jPos
	p.back[j] = iPos
}
