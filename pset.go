package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type Present int

const Present_Yes Present = 0

type Set []int

type SetKey struct {
	UnivSize int
	SetSize  int
	Delta    int
	Key      []byte
	prp      *PRP
}

// Go's % operator follows C semantics and can produce
// negative values if it's given a negative argument.
// We need an arithmetic mod operator.
func MathMod(x int, mod int) int {
	out := x % mod

	// TODO: This is not a constant-time operation.
	if out < 0 {
		out = out + mod
	}

	return out
}

func SetGen(src *rand.Rand, univSize int, setSize int) *SetKey {
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
	if l, err := src.Read(key); l != len(key) || err != nil {
		panic(err)
	}

	delta := src.Intn(univSize)
	prp, err := NewPRP(key, univSizeBits)
	if err != nil {
		panic(fmt.Errorf("Failed to create PRP: %s", err))
	}
	return &SetKey{univSize, setSize, delta, key, prp}
}

func SetGenWith(src *rand.Rand, univSize int, setSize int, val int) *SetKey {
	key := SetGen(src, univSize, setSize)
	key.Delta = 0

	// TODO: Implement this more efficiently.
	pos := src.Intn(setSize)
	key.Delta = MathMod(val-key.elemAt(pos), univSize)

	return key
}

func (key *SetKey) Punc(idx int) Set {
	out := make(Set, 0, key.SetSize-1)

	for i := 0; i < key.SetSize; i++ {
		elem := key.elemAt(i)
		if elem != idx {
			out = append(out, elem)
		}
	}
	return out
}

// Sample a random element of the set that is not equal to `idx`.
func (key *SetKey) RandomMemberExcept(randSource *rand.Rand, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := key.elemAt(randSource.Intn(key.SetSize))
		if val != idx {
			return val
		}
	}
}

func (key *SetKey) Eval() Set {
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
		out = append(out, key.elemAt(i))
	}

	return out
}

func (key *SetKey) InSet(idx int) bool {
	return key.prp.Invert(MathMod(idx-key.Delta, key.UnivSize)) < key.SetSize
}

func (key *SetKey) elemAt(pos int) int {
	elem := key.prp.Eval(pos)
	return MathMod(int(elem)+key.Delta, key.UnivSize)
}
