package boosted

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"

	"github.com/dimakogan/boosted-pir/psetggm"
)

type Present int

const Present_Yes Present = 0

type Set []int

type SetKey struct {
	id, shift uint32
}

type BaseGenerator interface {
	Eval(seed []byte, elems []int)
	Punc(seed []byte, pos int) []byte
	EvalPunctured(pset []byte, hole int, elems []int)
}

type PuncturableSet struct {
	SetKey
	univSize, setSize int
	seed              [16]byte
	elems             Set
}

type PuncturedSet struct {
	UnivSize, SetSize int
	Keys              []byte
	Hole              int
	Shift             uint32
}

type SetGenerator struct {
	baseGen           BaseGenerator
	num               uint32
	idGen             cipher.Block
	univSize, setSize int

	exists       []uint8
	existsMarker uint8
}

func NewSetGenerator(masterKey []byte, startId uint32, univSize int, setSize int) SetGenerator {
	aes, err := aes.NewCipher(masterKey)
	if err != nil {
		panic(err)
	}

	return SetGenerator{
		baseGen:      psetggm.NewGGMSetGeneratorC(univSize, setSize),
		num:          startId,
		idGen:        aes,
		univSize:     univSize,
		setSize:      setSize,
		exists:       make([]uint8, univSize),
		existsMarker: 0,
	}
}

func (gen *SetGenerator) Gen(pset *PuncturableSet) {
	gen.gen(pset)

	block := make([]byte, 16)
	out := make([]byte, 16)
	block[0] = 0xBB
	binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	gen.idGen.Encrypt(out, block)
	pset.shift = binary.LittleEndian.Uint32(out) % uint32(gen.setSize)

	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	}
}

func (gen *SetGenerator) GenWith(val int) (pset PuncturableSet) {
	gen.gen(&pset)

	block := make([]byte, 16)
	seed := make([]byte, 16)
	block[0] = 0xBB
	binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	gen.idGen.Encrypt(seed, block)
	pos := binary.LittleEndian.Uint64(seed) % uint64(gen.setSize)
	pset.shift = uint32(MathMod(val-pset.elems[pos], gen.univSize))

	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	}

	return pset
}

func (gen *SetGenerator) gen(pset *PuncturableSet) {
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != pset.setSize {
		pset.elems = make([]int, pset.setSize)
	}
	var block [16]byte

	for {
		block[0] = 0xAA
		binary.LittleEndian.PutUint32(block[1:], uint32(gen.num))
		pset.id = gen.num
		gen.num++

		gen.idGen.Encrypt(pset.seed[:], block[:])
		gen.baseGen.Eval(pset.seed[:], pset.elems)

		if gen.distinct2(pset.elems) {
			return
		}
	}
}

func (gen *SetGenerator) Eval(key SetKey) PuncturableSet {
	pset := PuncturableSet{
		SetKey:   key,
		univSize: gen.univSize,
		setSize:  gen.setSize,
		elems:    make([]int, gen.setSize)}

	var block [16]byte

	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], key.id)

	gen.idGen.Encrypt(pset.seed[:], block[:])
	gen.baseGen.Eval(pset.seed[:], pset.elems)

	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + key.shift) % uint32(gen.univSize))
	}

	return pset
}

func (gen *SetGenerator) Punc(pset PuncturableSet, idx int) PuncturedSet {
	for pos, elem := range pset.elems {
		if elem == idx {
			return PuncturedSet{
				UnivSize: pset.univSize,
				SetSize:  pset.setSize - 1,
				Hole:     pos,
				Shift:    pset.shift,
				Keys:     gen.baseGen.Punc(pset.seed[:], pos)}
		}
	}
	panic(fmt.Sprintf("Failed to find idx: %d in pset: %v", idx, pset.elems))
}

func (pset *PuncturedSet) Eval() Set {
	baseGen := psetggm.NewGGMSetGeneratorC(pset.UnivSize, pset.SetSize+1)
	elems := make([]int, pset.SetSize)
	baseGen.EvalPunctured(pset.Keys, pset.Hole, elems)
	for i := 0; i < len(elems); i++ {
		elems[i] = int((uint32(elems[i]) + pset.Shift) % uint32(pset.UnivSize))
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

func (gen *SetGenerator) distinct2(elems []int) bool {
	gen.existsMarker++
	if gen.existsMarker == 0 {
		for i := range gen.exists {
			gen.exists[i] = 0
		}
		gen.existsMarker++
	}
	for _, e := range elems {
		if gen.exists[e] == gen.existsMarker {
			return false
		}
		gen.exists[e] = gen.existsMarker
	}
	return true
}
