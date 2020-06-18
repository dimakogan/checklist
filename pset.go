package boosted

import (
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
	Delta    int
	Set      Set
}

type PuncSetKey struct {
	UnivSize int
	Delta    int
	Set      Set
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
	if univSize < setSize {
		panic("Set size too large.")
	}

	if univSize < 1 {
		panic("Universe size too small.")
	}

	// TODO: Implement this more efficiently
	out := make(Set)
	for len(out) < setSize {
		out[src.Intn(univSize)] = Present_Yes
	}

	delta := src.Intn(univSize)
	return &SetKey{univSize, delta, out}
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

func (key *SetKey) Punc(idx int) *PuncSetKey {
	puncAt := MathMod(idx-key.Delta, key.UnivSize)

	if _, okay := key.Set[puncAt]; !okay {
		panic("Can't puncture at this point!")
	}

	out := make(map[int]Present)
	for i := range key.Set {
		if i != puncAt {
			out[i] = Present_Yes
		}
	}

	return &PuncSetKey{key.UnivSize, key.Delta, out}
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
	return evalMap(key.UnivSize, key.Delta, key.Set)
}

func (key *PuncSetKey) Eval() Set {
	return evalMap(key.UnivSize, key.Delta, key.Set)
}

func evalMap(univSize int, delta int, m Set) Set {
	out := make(Set, len(m))

	for k := range m {
		out[MathMod(k+delta, univSize)] = Present_Yes
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
	return key.Eval().Has(idx)
}
