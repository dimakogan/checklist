package boosted

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
)

type ggmSet struct {
	key      []byte
	setSize  int
	univSize int
	height   int
	prg      cipher.Block
}

type ggmSetGenerator struct {
	src *rand.Rand
	prg cipher.Block
}

func NewGGMSetGenerator(src *rand.Rand) SetGenerator {
	prg, err := aes.NewCipher(zeroBlock)
	if err != nil {
		panic(fmt.Sprintf("Failed to create AES cipher: %s", err))
	}

	return &ggmSetGenerator{src: src, prg: prg}
}

func (g *ggmSetGenerator) SetGen(univSize int, setSize int) PuncturableSet {
	key := make([]byte, 16)
	height := int(math.Ceil(math.Log2(float64(setSize))))
	for {
		if l, err := g.src.Read(key); l != len(key) || err != nil {
			panic(err)
		}

		set := ggmSet{key: key, setSize: setSize, height: height, univSize: univSize,
			prg: g.prg,
		}

		if set.Eval().distinct() {
			return &set
		}
	}
}

func (set *ggmSet) Eval() Set {
	elems := make(Set, 1<<set.height)
	treeEvalAll(set.prg, set.key, set.height, set.univSize, elems)
	return elems[0:set.setSize]
}

func (set *ggmSet) Size() int {
	return set.setSize
}

func (set *ggmSet) Contains(idx int) bool {
	return set.findPos(idx) != -1
}

func (set *ggmSet) findPos(idx int) int {
	for i, v := range set.Eval() {
		if v == idx {
			return i
		}
	}
	return -1
}

func (set *ggmSet) ElemAt(pos int) int {
	return treeEval(set.prg, set.key, set.height, set.univSize, pos)
}

func treeEval(prg cipher.Block, key []byte, height int, univSize int, pos int) int {
	for ; height > 0; height-- {
		newKey := make([]byte, len(key))
		if pos < (1 << (height - 1)) {
			leftChild(prg, key, newKey)
		} else {
			rightChild(prg, key, newKey)
		}
		pos &^= (1 << (height - 1))
		key = newKey
	}

	return MathMod(int(binary.LittleEndian.Uint32(key)), univSize)

}

var zeroBlock = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func leftChild(prg cipher.Block, seed []byte, out []byte) {
	prg.Encrypt(out, seed)
	xorInto(out, seed)
}

// warning this is not thread safe since it mutates seed throughout (and reverts at the end)
// to avoid the extra copy
func rightChild(prg cipher.Block, seed []byte, out []byte) {
	seed[0] ^= 1
	prg.Encrypt(out, seed)
	xorInto(out, seed)
	seed[0] ^= 1
}

func treeEvalAll(prg cipher.Block, key []byte, height int, univSize int, out []int) {
	if height == 0 {
		out[0] = MathMod(int(binary.LittleEndian.Uint32(key)), univSize)
		return
	}
	nextKey := make([]byte, 16)
	leftChild(prg, key, nextKey)
	treeEvalAll(prg, nextKey, height-1, univSize, out[0:1<<(height-1)])
	rightChild(prg, key, nextKey)
	treeEvalAll(prg, nextKey, height-1, univSize, out[1<<(height-1):])
}

func (set *ggmSet) Punc(idx int) SuccinctSet {
	hole := set.findPos(idx)
	if hole < 0 {
		panic("Puncturing at non-existing element")
	}
	keys := make([][]byte, 0)
	key := set.key
	pos := hole
	for height := set.height; height > 0; height-- {
		pathKey := make([]byte, 16)
		copathKey := make([]byte, 16)
		if pos < (1 << (height - 1)) {
			leftChild(set.prg, key, pathKey)
			rightChild(set.prg, key, copathKey)
		} else {
			leftChild(set.prg, key, copathKey)
			rightChild(set.prg, key, pathKey)
		}
		keys = append(keys, copathKey)
		key = pathKey
		pos &^= (1 << (height - 1))
	}
	return &puncturedGGMSet{
		keys:     keys,
		hole:     hole,
		setSize:  set.setSize - 1,
		height:   set.height,
		univSize: set.univSize,
		prg:      set.prg,
	}
}

type puncturedGGMSet struct {
	keys     [][]byte
	hole     int
	setSize  int
	univSize int
	height   int
	prg      cipher.Block
}

func (set *puncturedGGMSet) Eval() Set {
	elems := make(Set, 1<<set.height)
	puncturedTreeEvalAll(set.prg, set.keys, set.hole, set.height, set.univSize, elems)
	return elems[0:set.setSize]
}

func puncturedTreeEvalAll(prg cipher.Block, keys [][]byte, hole int, height int, univSize int, out []int) {
	if height == 0 {
		return
	}
	if hole < (1 << (height - 1)) {
		puncturedTreeEvalAll(prg, keys[1:], hole, height-1, univSize, out[0:1<<(height-1)-1])
		treeEvalAll(prg, keys[0], height-1, univSize, out[1<<(height-1)-1:])
	} else {
		treeEvalAll(prg, keys[0], height-1, univSize, out[0:1<<(height-1)])
		puncturedTreeEvalAll(prg, keys[1:], hole-(1<<(height-1)), height-1, univSize, out[1<<(height-1):])
	}

}

func (set *puncturedGGMSet) elemAt(pos int) int {
	if pos >= set.hole {
		pos++
	}
	hole := set.hole
	height := set.height
	for {
		holeMsb := hole & (1<<height - 1)
		posMsb := pos & (1<<height - 1)
		hole &^= (1<<height - 1)
		pos &^= (1<<height - 1)
		height--

		if (holeMsb) != (posMsb) {
			break
		}
	}
	if height == 0 {
		panic("Cannot evaluate punctured set at punctured point")
	}
	return treeEval(set.prg, set.keys[set.height-height], height, set.univSize, pos)
}

func (set *puncturedGGMSet) Size() int {
	return set.setSize
}
