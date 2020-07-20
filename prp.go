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
	round1           cipher.Block
	round2           cipher.Block
	round3           cipher.Block
	blockLenBits     int
	prfBlockLenBytes int
	lastByteMask     byte

	// Preallocated buffers
	u []byte
	v []byte
	w []byte
	x []byte
	y []byte
}

func NewPRP(key []byte, blockLenBits int) (*PRP, error) {
	if blockLenBits%2 != 0 {
		return nil, fmt.Errorf("Block len must be even")
	}
	if blockLenBits > 32 {
		return nil, fmt.Errorf("Block len must be smaller than 32 bits")
	}

	prp := PRP{
		blockLenBits:     blockLenBits,
		prfBlockLenBytes: (blockLenBits/2-1)/8 + 1,
		lastByteMask:     byte((1 << ((blockLenBits / 2) % 8)) - 1),
		u:                make([]byte, aes.BlockSize),
		v:                make([]byte, aes.BlockSize),
		w:                make([]byte, aes.BlockSize),
		x:                make([]byte, aes.BlockSize),
		y:                make([]byte, aes.BlockSize),
	}

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

func (prp *PRP) Eval(in int) int {
	// Parse input to binary
	binary.LittleEndian.PutUint32(prp.u, uint32(in)&((1<<(prp.blockLenBits/2))-1))
	binary.LittleEndian.PutUint32(prp.v, uint32(in)>>(prp.blockLenBits/2))

	prp.prf(prp.round1, prp.w, prp.v)
	fastxor.Bytes(prp.w, prp.w, prp.u)

	prp.prf(prp.round2, prp.x, prp.w)
	fastxor.Bytes(prp.x, prp.x, prp.v)

	prp.prf(prp.round3, prp.y, prp.x)
	fastxor.Bytes(prp.y, prp.y, prp.w)

	return int(binary.LittleEndian.Uint32(prp.x) + binary.LittleEndian.Uint32(prp.y)<<(prp.blockLenBits/2))
}

func (prp *PRP) Invert(in int) int {
	// Parse input to binary
	binary.LittleEndian.PutUint32(prp.x, uint32(in)&((1<<(prp.blockLenBits/2))-1))
	binary.LittleEndian.PutUint32(prp.y, uint32(in)>>(prp.blockLenBits/2))

	prp.prf(prp.round3, prp.w, prp.x)
	fastxor.Bytes(prp.w, prp.w, prp.y)

	prp.prf(prp.round2, prp.v, prp.w)
	fastxor.Bytes(prp.v, prp.v, prp.x)

	prp.prf(prp.round1, prp.u, prp.v)
	fastxor.Bytes(prp.u, prp.u, prp.w)

	return int(binary.LittleEndian.Uint32(prp.u) + binary.LittleEndian.Uint32(prp.v)<<(prp.blockLenBits/2))
}

func (prp *PRP) prf(c cipher.Block, dst []byte, src []byte) {
	c.Encrypt(dst, src)
	// zero out the remainder of the AES block
	fastxor.Bytes(dst[prp.prfBlockLenBytes:], dst[prp.prfBlockLenBytes:], dst[prp.prfBlockLenBytes:])
	dst[prp.prfBlockLenBytes-1] &= prp.lastByteMask
}
