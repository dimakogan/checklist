package boosted

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/lukechampine/fastxor"
	"golang.org/x/crypto/hkdf"

	"crypto/sha256"
)

type PRP struct {
	round1       cipher.Block
	round2       cipher.Block
	round3       cipher.Block
	blockLenBits int
	mask         []byte
}

func NewPRP(key []byte, blockLenBits int) (*PRP, error) {
	if blockLenBits%2 != 0 {
		return nil, fmt.Errorf("Block len must be even")
	}
	if blockLenBits > 32 {
		return nil, fmt.Errorf("Block len must be smaller than 32 bits")
	}

	prp := PRP{
		blockLenBits: blockLenBits,
		mask:         make([]byte, aes.BlockSize),
	}

	var mask uint32
	mask = (1 << (blockLenBits / 2)) - 1
	binary.LittleEndian.PutUint32(prp.mask, mask)

	hash := sha256.New
	hkdf := hkdf.New(hash, key, nil, nil)
	k1 := make([]byte, 16)
	k2 := make([]byte, 16)
	k3 := make([]byte, 16)
	if _, err := io.ReadFull(hkdf, k1); err != nil {
		panic(err)
	}
	if _, err := io.ReadFull(hkdf, k2); err != nil {
		panic(err)
	}
	if _, err := io.ReadFull(hkdf, k3); err != nil {
		panic(err)
	}

	var err error
	prp.round1, err = aes.NewCipher(k1)
	if err != nil {
		return nil, errors.New("failed to create AES block cipher")
	}
	prp.round2, err = aes.NewCipher(k2)
	if err != nil {
		return nil, errors.New("failed to create AES block cipher")
	}
	prp.round3, err = aes.NewCipher(k3)
	if err != nil {
		return nil, errors.New("failed to create AES block cipher")
	}

	return &prp, nil
}

func (prp *PRP) Eval(in uint32) uint32 {
	// Parse input to binary
	u := make([]byte, aes.BlockSize)
	v := make([]byte, aes.BlockSize)
	binary.LittleEndian.PutUint32(u, in&((1<<(prp.blockLenBits/2))-1))
	binary.LittleEndian.PutUint32(v, in>>(prp.blockLenBits/2))

	w := make([]byte, aes.BlockSize)
	prp.prf(prp.round1, w, v)
	fastxor.Bytes(w, w, u)

	x := make([]byte, aes.BlockSize)
	prp.prf(prp.round2, x, w)
	fastxor.Bytes(x, x, v)

	y := make([]byte, aes.BlockSize)
	prp.prf(prp.round2, y, x)
	fastxor.Bytes(y, y, w)

	return binary.LittleEndian.Uint32(x) + binary.LittleEndian.Uint32(y)<<(prp.blockLenBits/2)
}

func (prp *PRP) Invert(y int) int {
	return 0
}

func (prp *PRP) prf(c cipher.Block, dst []byte, src []byte) {
	c.Encrypt(dst, src)
	for i := (prp.blockLenBits/2-1)/8 + 1; i < aes.BlockSize; i++ {
		dst[i] = 0
	}
	lastByteMask := byte((1 << ((prp.blockLenBits / 2) % 8)) - 1)

	dst[(prp.blockLenBits/2-1)/8] &= lastByteMask
}
