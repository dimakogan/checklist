package pir

import (
  "encoding/binary"
  "log"
  "crypto/rand"
)

type cryptoSource struct {}

func (s cryptoSource) Int63() int64 {
  var mask uint64 = 0x7fffffffffffffff
  return int64(s.Uint64() & mask)
}

func (cryptoSource) Uint64() uint64 {
  var buf [8]byte
  _, err := rand.Read(buf[:])
  if err != nil {
    log.Fatal("rand.Read failed")
  }

  return binary.LittleEndian.Uint64(buf[:])
}

func (cryptoSource) Seed(int64) {
  log.Fatal("Not implemented.")
}

