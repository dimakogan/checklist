package main

import (
  "encoding/binary"
  "log"
  "sort"
)

func partialHashTo32(hash PartialHash) uint32 {
  if len(hash) != PartialHashLen {
    log.Fatal("Invalid partial hash len")
  }

  return binary.LittleEndian.Uint32(hash)
}

func partialHashInts() []uint32 {
  hashes := fetchPartialHashes()
  if len(hashes) == 0 {
    log.Fatal("Should get more hashes")
  }
  hashInts := make([]uint32, len(hashes))
  for i,h := range hashes {
    hashInts[i] = partialHashTo32(h)
  }

  sort.Slice(hashInts[:], func(i int, j int) bool { return hashInts[i] < hashInts[j] })

  log.Printf("hashes = %v", hashInts)
  return hashInts
}

func riceDeltas(hashInts []uint32) []byte {
  out := make([]byte, 4*(len(hashInts) - 1))
  for i,_ := range hashInts {
    log.Printf("Hash[%v] = %v", i, hashInts[i])

    if i == 0 {
      continue
    }

    delta := hashInts[i] - hashInts[i-1]

    if delta & (0x1 << 31) > 0 {
      panic("Oh no!")
    }

    out[4*(i-1) + 0] = byte((delta & (0x7F)) << 1)
    out[4*(i-1) + 1] = byte((delta & (0xFF << 7)) >> 7)
    out[4*(i-1) + 2] = byte((delta & (0xFF << 15)) >> 15)
    out[4*(i-1) + 3] = byte((delta & (0xFF << 23)) >> 23)
  }
  return out
}

func riceEncodedHashes() *RiceDeltaEncoding {
  hashInts := append([]uint32{0}, partialHashInts()...)
  rice := new(RiceDeltaEncoding)
  rice.RiceParameter = 31
  rice.NumEntries = 0
  if len(hashInts) > 1 {
    rice.NumEntries = int32(len(hashInts) - 1)
  }
  rice.FirstValue = int64(hashInts[0])
  rice.EncodedData = riceDeltas(hashInts)
  log.Printf("FirstValue = %v", rice.FirstValue)
  log.Printf("Encoded = %v", rice.EncodedData)

  return rice
}

