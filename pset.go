package boosted

import "math/rand"

type Present int

const Present_Yes Present = 0

type Set []int

type SuccinctSet interface {
	Size() int
	Eval() Set
}

type PuncturableSet interface {
	Size() int
	Eval() Set
	Contains(idx int) bool
	ElemAt(pos int) int
	Punc(idx int) SuccinctSet
}

type SetGenerator interface {
	SetGen(univSize int, setSize int) PuncturableSet
}

type shiftedSet struct {
	baseSet              SuccinctSet
	baseSetAsPuncturable PuncturableSet
	delta                int
	univSize             int
}

func SetGenWith(g SetGenerator, src *rand.Rand, univSize int, setSize int, val int) PuncturableSet {
	baseSet := g.SetGen(univSize, setSize)

	// TODO: Implement this more efficiently.
	pos := src.Intn(setSize)

	return &shiftedSet{
		baseSet:              baseSet,
		baseSetAsPuncturable: baseSet,
		delta:                MathMod(val-baseSet.ElemAt(pos), univSize),
		univSize:             univSize,
	}
}

func SetGen(g SetGenerator, src *rand.Rand, univSize int, setSize int) PuncturableSet {
	return SetGenWith(g, src, univSize, setSize, src.Intn(univSize))
}

func (ss *shiftedSet) Eval() Set {
	elems := ss.baseSet.Eval()
	for i := 0; i < len(elems); i++ {
		elems[i] = MathMod(elems[i]+ss.delta, ss.univSize)
	}
	return elems
}

func (ss *shiftedSet) Contains(idx int) bool {
	return ss.baseSetAsPuncturable.Contains(MathMod(idx-ss.delta, ss.univSize))
}

func (ss *shiftedSet) ElemAt(pos int) int {
	return MathMod(ss.baseSetAsPuncturable.ElemAt(pos)+ss.delta, ss.univSize)
}
func (ss *shiftedSet) Punc(idx int) SuccinctSet {
	return &shiftedSet{
		baseSet:  ss.baseSetAsPuncturable.Punc(MathMod(idx-ss.delta, ss.univSize)),
		univSize: ss.univSize,
		delta:    ss.delta,
	}
}

func (ss *shiftedSet) Size() int {
	return ss.baseSet.Size()
}

func (s Set) ElemAt(pos int) int {
	return s[pos]
}

func (s Set) Eval() Set {
	out := make(Set, len(s))
	copy(out, s)
	return out
}

func (s Set) Contains(idx int) bool {
	for _, elem := range s {
		if elem == idx {
			return true
		}
	}
	return false
}

func (s Set) Size() int {
	return len(s)
}

func (s Set) Punc(idx int) SuccinctSet {
	elems := make(Set, 0, len(s)-1)
	for _, elem := range s {
		if elem != idx {
			elems = append(elems, elem)
		}
	}
	return elems
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

func (set Set) distinct() bool {
	elemsSet := make(map[int]bool, len(set))
	for i := 0; i < len(set); i++ {
		elem := set[i]
		if _, ok := elemsSet[elem]; ok {
			return false
		}
		elemsSet[elem] = true
	}
	return len(elemsSet) == len(set)
}
