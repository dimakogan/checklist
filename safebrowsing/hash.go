package safebrowsing

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"sort"
)

const PartialHashLen = 4

type PartialHash = []byte

func ComputeHash(url []byte) (partial PartialHash, full []byte) {
	hash := sha256.New()
	hash.Write(url)
	full = hash.Sum(nil)
	return full[0:PartialHashLen], full
}

func PartialHashTo32(hash PartialHash) uint32 {
	if len(hash) != PartialHashLen {
		log.Fatal("Invalid partial hash len")
	}

	return binary.LittleEndian.Uint32(hash)
}

func PartialHashesToInts(hashes []PartialHash) []uint32 {
	if len(hashes) == 0 {
		log.Fatal("Should get more hashes")
	}
	hashInts := make([]uint32, len(hashes))
	for i, h := range hashes {
		hashInts[i] = PartialHashTo32(h)
	}

	sort.Slice(hashInts[:], func(i int, j int) bool { return hashInts[i] < hashInts[j] })

	log.Printf("hashes = %v", hashInts)
	return hashInts
}
