package boosted

import (
	"fmt"
	"math"
	"math/rand"
)

type Present int

const Present_Yes Present = 0

// This represents a set. If if key `k` is in the map,
// it means that `k` is in the set. We could use `bool` types
// here instead, but then a key can be defined and its value
// is `false`, which is confusing.
type Set map[int]Present

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

func (s Set) RandomMember(src *rand.Rand) int {
	keys := make([]int, len(s))
	i := 0
	for k := range s {
		keys[i] = k
		i += 1
	}

	choose := src.Intn(len(keys))
	return keys[choose]
}

func (s Set) Has(elm int) bool {
	_, okay := s[elm]
	return okay
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
	set := key.Eval()
	choose := set.RandomMember(src)
	key.Delta = MathMod(val-choose, univSize)

	return key
}

func (key *SetKey) Shift(amount int) {
	key.Delta = MathMod(key.Delta+amount, key.UnivSize)
}

func (key *SetKey) Punc(idx int) Set {
	set := key.Eval()

	if _, okay := set[idx]; !okay {
		panic("Can't puncture at this point!")
	}

	delete(set, idx)

	return set
}

func (key *SetKey) RandomMember(randSource *rand.Rand) int {
	return key.Eval().RandomMember(randSource)
}

// Sample a random element of the set that is not equal to `idx`.
func (key *SetKey) RandomMemberExcept(randSource *rand.Rand, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		val := key.Eval().RandomMember(randSource)
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
	out := make(Set, key.SetSize)

	for i := 0; i < key.SetSize; i++ {
		elem := key.prp.Eval(i)
		out[MathMod(int(elem)+key.Delta, key.UnivSize)] = Present_Yes
	}

	return out
}

// Given set key `key`, an element of the universe `idx`, and a slice
// `deltas` of shift values, find a `j` in `deltas` such that `idx` is
// in the set `key.Shift(deltas[j])`.
//
// Returns -1 if no such value exists.
func (key *SetKey) FindShift(idx int, deltas []int) int {
	set := key.Eval()

	for j, delta := range deltas {
		shift := MathMod(idx-delta, key.UnivSize)
		if set.Has(shift) {
			return j
		}
	}

	return -1
}

func (key *SetKey) InSet(idx int) bool {
	return key.prp.Invert(MathMod(idx-key.Delta, key.UnivSize)) < key.SetSize
}
