package boosted

import (
	"encoding/binary"
	"io"
	"math/rand"
)

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
	SetGenAndEval(univSize int, setSize int) (PuncturableSet, Set)
}

type shiftedSetGenerator struct {
	SetGenerator
	src *rand.Rand
}

type ShiftedSet struct {
	BaseSet              SuccinctSet
	baseSetAsPuncturable PuncturableSet
	Delta                int
	UnivSize             int
}

type NewGeneratorFunc func(io.Reader) SetGenerator

func NewSetGenerator(
	newGen NewGeneratorFunc,
	masterKey []byte) *shiftedSetGenerator {

	seed := int64(binary.LittleEndian.Uint64(masterKey))
	src1 := rand.New(rand.NewSource(seed))
	src2 := rand.New(rand.NewSource(seed))

	return &shiftedSetGenerator{
		SetGenerator: newGen(src1),
		src:          src2,
	}
}

func (g shiftedSetGenerator) GenWith(univSize int, setSize int, val int) PuncturableSet {
	baseSet := g.SetGenerator.SetGen(univSize, setSize)
	pos := g.src.Intn(setSize)

	return &ShiftedSet{
		BaseSet:              baseSet,
		baseSetAsPuncturable: baseSet,
		Delta:                MathMod(val-baseSet.ElemAt(pos), univSize),
		UnivSize:             univSize,
	}
}

func (g shiftedSetGenerator) SetGen(univSize int, setSize int) PuncturableSet {
	pset, _ := g.SetGenAndEval(univSize, setSize)
	return pset
}

func (g shiftedSetGenerator) SetGenAndEval(univSize int, setSize int) (PuncturableSet, Set) {
	baseSet, elems := g.SetGenerator.SetGenAndEval(univSize, setSize)
	ss := ShiftedSet{
		BaseSet:              baseSet,
		baseSetAsPuncturable: baseSet,
		Delta:                g.src.Intn(univSize),
		UnivSize:             univSize,
	}

	for i := 0; i < len(elems); i++ {
		elems[i] = int(uint32(elems[i]+ss.Delta) % uint32(ss.UnivSize))
	}
	return &ss, elems
}

func (ss *ShiftedSet) Eval() Set {
	elems := ss.BaseSet.Eval()
	for i := 0; i < len(elems); i++ {
		elems[i] = int(uint32(elems[i]+ss.Delta) % uint32(ss.UnivSize))
	}
	return elems
}

func (ss *ShiftedSet) Contains(idx int) bool {
	return ss.baseSetAsPuncturable.Contains(MathMod(idx-ss.Delta, ss.UnivSize))
}

func (ss *ShiftedSet) ElemAt(pos int) int {
	return MathMod(ss.baseSetAsPuncturable.ElemAt(pos)+ss.Delta, ss.UnivSize)
}
func (ss *ShiftedSet) Punc(idx int) SuccinctSet {
	return &ShiftedSet{
		BaseSet:  ss.baseSetAsPuncturable.Punc(MathMod(idx-ss.Delta, ss.UnivSize)),
		UnivSize: ss.UnivSize,
		Delta:    ss.Delta,
	}
}

func (ss *ShiftedSet) Size() int {
	return ss.BaseSet.Size()
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
	return true
}
