package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type prpSetKey struct {
	UnivSize int
	SetSize  int
	Key      []byte
	prp      *PRP
}

type prpSetGenerator struct {
	src *rand.Rand
}

func NewPrpSetGenerator(src *rand.Rand) *prpSetGenerator {
	return &prpSetGenerator{
		src: src,
	}
}

func (g prpSetGenerator) SetGen(univSize int, setSize int) SetKey {
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
	return &prpSetKey{univSize, setSize, key, prp}
}

func (key *prpSetKey) Size() int {
	return key.SetSize
}

func (key *prpSetKey) Punc(idx int) SetKey {
	out := make(Set, 0, key.SetSize-1)

	for i := 0; i < key.SetSize; i++ {
		elem := key.ElemAt(i)
		if elem != idx {
			out = append(out, elem)
		}
	}
	return out
}

func (key *prpSetKey) Eval() Set {
	// This is a workaround for prp not being serialized yet
	var err error
	if key.prp == nil {
		key.prp, err = NewPRP(key.Key, int(math.Log2(float64(key.UnivSize))))
		if err != nil {
			panic(fmt.Errorf("Failed to create PRP: %s", err))
		}
	}
	out := make(Set, 0, key.SetSize)

	for i := 0; i < key.SetSize; i++ {
		out = append(out, key.ElemAt(i))
	}

	return out
}

func (key *prpSetKey) InSet(idx int) bool {
	return key.prp.Invert(idx) < key.SetSize
}

func (key *prpSetKey) ElemAt(pos int) int {
	return key.prp.Eval(pos)
}
