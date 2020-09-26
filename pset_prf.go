package boosted

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
)

type prfSet struct {
	UnivSize int
	SetSize  int
	Key      []byte
	prf      cipher.Block
}

type prfSetGenerator struct {
	src io.Reader
}

func NewPRFSetGenerator(src io.Reader) SetGenerator {
	return &prfSetGenerator{
		src: src,
	}
}

func (g prfSetGenerator) SetGen(univSize int, setSize int) PuncturableSet {
	pset, _ := g.SetGenAndEval(univSize, setSize)
	return pset
}

func (g prfSetGenerator) SetGenAndEval(univSize int, setSize int) (PuncturableSet, Set) {
	if univSize < setSize {
		panic("Set size too large.")
	}

	key := make([]byte, 16)
	for {
		if l, err := g.src.Read(key); l != len(key) || err != nil {
			panic(err)
		}

		prf, err := aes.NewCipher(key)
		if err != nil {
			panic(fmt.Sprintf("Failed to create AES cipher: %s", err))
		}
		pset := prfSet{univSize, setSize, key, prf}
		if set := pset.Eval(); set.distinct() {
			return &pset, set
		}
	}
	return nil, nil
}

func (set *prfSet) Size() int {
	return set.SetSize
}

func (set *prfSet) Punc(idx int) SuccinctSet {
	out := make(Set, 0, set.SetSize-1)

	for i := 0; i < set.SetSize; i++ {
		elem := set.ElemAt(i)
		if elem != idx {
			out = append(out, elem)
		}
	}
	return out
}

func (set *prfSet) Eval() Set {
	// This is a workaround for prf not being serialized yet
	var err error
	if set.prf == nil {
		set.prf, err = aes.NewCipher(set.Key)
		if err != nil {
			panic(fmt.Sprintf("Failed to create AES cipher: %s", err))
		}
	}
	out := make(Set, 0, set.SetSize)

	for i := 0; i < set.SetSize; i++ {
		out = append(out, set.ElemAt(i))
	}

	return out
}

func (set *prfSet) Contains(idx int) bool {
	return set.findPos(idx) != -1
}

func (set *prfSet) findPos(idx int) int {
	for i, v := range set.Eval() {
		if v == idx {
			return i
		}
	}
	return -1
}

func (set *prfSet) ElemAt(pos int) int {
	in := make([]byte, aes.BlockSize)
	binary.LittleEndian.PutUint32(in, uint32(pos))
	out := make([]byte, aes.BlockSize)
	set.prf.Encrypt(out, in)
	return int(binary.LittleEndian.Uint32(out) % uint32(set.UnivSize))
}
