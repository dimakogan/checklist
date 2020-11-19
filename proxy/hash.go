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

  sort.Slice(hashInts[:], func(i int, j int) bool { return i < j })

  return hashInts
}

func riceDeltas(hashInts []uint32) []byte {
  out := make([]byte, 4*len(hashInts))
  for i,_ := range hashInts {
    if i == 0 {
      continue
    }

    offset := i-1
    delta := hashInts[i] - hashInts[i-1]

    if delta & (0x1 << 31) > 0 {
      panic("Oh no!")
    }

    binary.LittleEndian.PutUint32(out[4*offset:], delta)
  }
  return out
}

func riceEncodedHashes() *RiceDeltaEncoding {
  hashInts := partialHashInts()
  rice := new(RiceDeltaEncoding)
  rice.RiceParameter = 31
  rice.NumEntries = 0
  rice.FirstValue = int64(hashInts[0])
  rice.EncodedData = riceDeltas(hashInts)

  return rice
}

