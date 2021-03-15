package pir

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"log"
	"math/big"
	"sync"
)

type PRGKey [aes.BlockSize]byte

const bufSize = 8192

// Copied from: https://github.com/henrycg/prio/blob/master/utils/rand.go
// We use the AES-CTR to generate pseudo-random  numbers using a
// stream cipher. Go's native rand.Reader is extremely slow because
// it makes tons of system calls to generate a small number of
// pseudo-random bytes.
//
// We pay the overhead of using a sync.Mutex to synchronize calls
// to AES-CTR, but this is relatively cheap.
type PRGReader struct {
	Key      PRGKey
	stream   cipher.Stream
	prgMutex sync.Mutex
}

type BufPRGReader struct {
	Key    PRGKey
	stream *bufio.Reader
}

func NewPRG(key *PRGKey) *PRGReader {
	out := new(PRGReader)
	out.Key = *key

	var err error
	var iv [aes.BlockSize]byte

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	out.stream = cipher.NewCTR(block, iv[:])
	return out
}

func RandomPRGKey() *PRGKey {
	var key PRGKey
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}

	return &key
}

func (s *PRGReader) Read(p []byte) (int, error) {
	s.prgMutex.Lock()
	defer s.prgMutex.Unlock()
	if len(p) < aes.BlockSize {
		var buf [aes.BlockSize]byte
		s.stream.XORKeyStream(buf[:], buf[:])
		copy(p[:], buf[:])
	} else {
		s.stream.XORKeyStream(p, p)
	}

	return len(p), nil
}

func NewBufPRG(prg *PRGReader) *BufPRGReader {
	out := new(BufPRGReader)
	out.Key = prg.Key
	out.stream = bufio.NewReaderSize(prg, bufSize)
	return out
}

func (b *BufPRGReader) RandInt(n int) int {
	max := new(big.Int)
	max.SetInt64(int64(n))
	out, err := rand.Int(b.stream, max)
	if err != nil {
		// TODO: Replace this with non-absurd error handling.
		panic("Catastrophic randomness failure!")
	}

	return int(out.Int64())
}

func (b *BufPRGReader) Uint64() uint64 {
	var buf [8]byte
	_, err := b.stream.Read(buf[:])
	if err != nil {
		log.Fatal("rand.Read failed")
	}

	return binary.LittleEndian.Uint64(buf[:])
}

func (b *BufPRGReader) Int63() int64 {
	var mask uint64 = 0x7fffffffffffffff
	return int64(b.Uint64() & mask)
}

func (b *BufPRGReader) Seed(int64) {
	log.Fatal("Not implemented.")
}
