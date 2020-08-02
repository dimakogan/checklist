package boosted

import (
	"crypto/aes"
	"encoding/binary"
	"math"
	"math/rand"
)

type psetGGM struct {
	key      []byte
	setSize  int
	univSize int
	height   int
}

type ggmSetGenerator struct {
	src *rand.Rand
}

func NewGGMSetGenerator(src *rand.Rand) SetGenerator {
	return &ggmSetGenerator{src: src}
}

func (g *ggmSetGenerator) SetGen(univSize int, setSize int) SetKey {
	key := make([]byte, 16)
	height := int(math.Ceil(math.Log2(float64(setSize))))
	for {
		if l, err := g.src.Read(key); l != len(key) || err != nil {
			panic(err)
		}
		set := psetGGM{key: key, setSize: setSize, height: height, univSize: univSize}
		elems := set.Eval()

		elemsSet := make(map[int]bool, setSize)
		for i := 0; i < setSize; i++ {
			elem := elems[i]
			if _, ok := elemsSet[elem]; ok {
				break
			}
			elemsSet[elem] = true
		}
		if len(elemsSet) == set.setSize {
			return &set
		}
	}
}

func (set *psetGGM) Eval() Set {
	elems := make(Set, 1<<set.height)
	treeEvalAll(set.key, set.height, set.univSize, elems)
	return elems[0:set.setSize]
}

func (set *psetGGM) Size() int {
	return set.setSize
}

func (set *psetGGM) InSet(idx int) bool {
	return set.findPos(idx) != -1
}

func (set *psetGGM) findPos(idx int) int {
	for i, v := range set.Eval() {
		if v == idx {
			return i
		}
	}
	return -1
}

func (set *psetGGM) ElemAt(pos int) int {
	return treeEval(set.key, set.height, set.univSize, pos)
}

var zeroBlock = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var oneBlock = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

func treeEval(key []byte, height int, univSize int, pos int) int {
	for ; height > 0; height-- {
		prg, err := aes.NewCipher(key)
		if err != nil {
			panic("failed to create AES block cipher")
		}
		newKey := make([]byte, len(key))
		if pos < (1 << (height - 1)) {
			prg.Encrypt(newKey, zeroBlock)
		} else {
			prg.Encrypt(newKey, oneBlock)
		}
		pos &^= (1 << (height - 1))
		key = newKey
	}

	return MathMod(int(binary.LittleEndian.Uint32(key)), univSize)

}

func treeEvalAll(key []byte, height int, univSize int, out []int) {
	if height == 0 {
		out[0] = MathMod(int(binary.LittleEndian.Uint32(key)), univSize)
		return
	}

	prg, err := aes.NewCipher(key)
	if err != nil {
		panic("failed to create AES block cipher")
	}

	nextKey := make([]byte, 16)
	prg.Encrypt(nextKey, zeroBlock)
	treeEvalAll(nextKey, height-1, univSize, out[0:1<<(height-1)])
	prg.Encrypt(nextKey, oneBlock)
	treeEvalAll(nextKey, height-1, univSize, out[1<<(height-1):])
}

func (set *psetGGM) Punc(idx int) SetKey {
	hole := set.findPos(idx)
	if hole < 0 {
		panic("Puncturing at non-existing element")
	}
	keys := make([][]byte, 0)
	key := set.key
	pos := hole
	for height := set.height; height > 0; height-- {
		prg, err := aes.NewCipher(key)
		if err != nil {
			panic("failed to create AES block cipher")
		}
		pathKey := make([]byte, 16)
		copathKey := make([]byte, 16)
		if pos < (1 << (height - 1)) {
			prg.Encrypt(copathKey, oneBlock)
			prg.Encrypt(pathKey, zeroBlock)
		} else {
			prg.Encrypt(copathKey, zeroBlock)
			prg.Encrypt(pathKey, oneBlock)
		}
		keys = append(keys, copathKey)
		key = pathKey
		pos &^= (1 << (height - 1))
	}
	return &puncturedSetGGM{
		keys:     keys,
		hole:     hole,
		setSize:  set.setSize - 1,
		height:   set.height,
		univSize: set.univSize,
	}
}

type puncturedSetGGM struct {
	keys     [][]byte
	hole     int
	setSize  int
	univSize int
	height   int
}

func (set *puncturedSetGGM) Eval() Set {
	elems := make(Set, 1<<set.height)
	puncturedTreeEvalAll(set.keys, set.hole, set.height, set.univSize, elems)
	return elems[0:set.setSize]
}

func puncturedTreeEvalAll(keys [][]byte, hole int, height int, univSize int, out []int) {
	if height == 0 {
		return
	}
	if hole < (1 << (height - 1)) {
		puncturedTreeEvalAll(keys[1:], hole, height-1, univSize, out[0:1<<(height-1)-1])
		treeEvalAll(keys[0], height-1, univSize, out[1<<(height-1)-1:])
	} else {
		treeEvalAll(keys[0], height-1, univSize, out[0:1<<(height-1)])
		puncturedTreeEvalAll(keys[1:], hole-(1<<(height-1)), height-1, univSize, out[1<<(height-1):])
	}

}

func (set *puncturedSetGGM) ElemAt(pos int) int {
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
		return -1
	}
	return treeEval(set.keys[set.height-height], height, set.univSize, pos)
}

func (set *puncturedSetGGM) InSet(pos int) bool {
	panic("Not implemented")
	return false
}

func (set *puncturedSetGGM) Punc(pos int) SetKey {
	panic("Cannot double puncture a set")
	return nil
}

func (set *puncturedSetGGM) Size() int {
	return set.setSize
}
