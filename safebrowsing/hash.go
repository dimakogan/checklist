package safebrowsing

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"log"
	"sort"
	"strings"
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

func ReadBlockedURLs(f io.ReadCloser) (keys []uint32, values [][]byte, err error) {
	defer f.Close()
	scanner := bufio.NewScanner(f)
	keys = make([]uint32, 0)
	values = make([][]byte, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		partial, full := ComputeHash([]byte(line))
		keys = append(keys, binary.LittleEndian.Uint32(partial))
		values = append(values, full)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return keys, values, nil
}
