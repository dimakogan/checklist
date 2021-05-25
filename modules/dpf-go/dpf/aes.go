// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpf

// defined in asm_amd64.s
// extern xor16
func xor16(dst, a, b *byte)
func encryptAes128(xk *uint32, dst, src *byte)
func aes128MMO(xk *uint32, dst, src *byte)
func expandKeyAsm(key *byte, enc *uint32)

type aesPrf struct {
	enc []uint32
}

func newCipher(key []byte) (*aesPrf, error) {
	n := 11*4
	c := aesPrf{make([]uint32, n)}
	expandKeyAsm(&key[0], &c.enc[0])
	return &c, nil
}

func (c *aesPrf) BlockSize() int { return 16 }

func (c *aesPrf) Encrypt(dst, src []byte) {
	encryptAes128(&c.enc[0], &dst[0], &src[0])
}



