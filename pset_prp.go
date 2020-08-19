package boosted

import (
	"fmt"
	"io"
	"math"
)

type prpSet struct {
	UnivSize int
	SetSize  int
	Key      []byte
	prp      *PRP
}

type prpSetGenerator struct {
	src io.Reader
}

func NewPRPSetSetGenerator(src io.Reader) SetGenerator {
	return &prpSetGenerator{
		src: src,
	}
}

func (g prpSetGenerator) SetGen(univSize int, setSize int) PuncturableSet {
	if univSize < 4 {
		panic("Universe size must be at least 4.")
	}

	if univSize < setSize {
		panic("Set size too large.")
	}

	if (univSize & (univSize - 1)) != 0 {
		panic("Universe size is not a power of 2.")
	}

	univSizeBits := int(math.Log2(float64(univSize)))
	if univSizeBits%2 != 0 {
		panic("Universe size is not an EVEN power of 2.")
	}

	key := make([]byte, 16)
	if l, err := g.src.Read(key); l != len(key) || err != nil {
		panic(err)
	}

	prp, err := NewPRP(key, univSizeBits)
	if err != nil {
		panic(fmt.Errorf("Failed to create PRP: %s", err))
	}
	return &prpSet{univSize, setSize, key, prp}
}

func (set *prpSet) Size() int {
	return set.SetSize
}

func (set *prpSet) Punc(idx int) SuccinctSet {
	out := make(Set, 0, set.SetSize-1)

	for i := 0; i < set.SetSize; i++ {
		elem := set.ElemAt(i)
		if elem != idx {
			out = append(out, elem)
		}
	}
	return out
}

func (set *prpSet) Eval() Set {
	// This is a workaround for prp not being serialized yet
	var err error
	if set.prp == nil {
		set.prp, err = NewPRP(set.Key, int(math.Log2(float64(set.UnivSize))))
		if err != nil {
			panic(fmt.Errorf("Failed to create PRP: %s", err))
		}
	}
	out := make(Set, 0, set.SetSize)

	for i := 0; i < set.SetSize; i++ {
		out = append(out, set.ElemAt(i))
	}

	return out
}

func (set *prpSet) Contains(idx int) bool {
	return set.prp.Invert(idx) < set.SetSize
}

func (set *prpSet) ElemAt(pos int) int {
	return set.prp.Eval(pos)
}
