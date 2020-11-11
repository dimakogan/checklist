package main

import (
	"crypto/sha256"
)

//const evilHost = "testsafebrowsing.appspot.com/s/phishing.html"
const evilHost = "https://testsafebrowsing.appspot.com/apiv4/IOS/MALWARE/URL/"

const PartialHashLen = 4
type PartialHash = []byte

func evilHash() []byte {
	hash := sha256.New()
	hash.Write([]byte(evilHost))
  return hash.Sum(nil)
}

func getPartialHashes() []PartialHash {
  return []PartialHash{ evilHash()[0:PartialHashLen] }
}

// This is the function that will make the PIR query.
// It takes as input an index `idx` and outputs the
// full SHA256 hash of the URL corresponding to that index.
func queryForHash(idx int) []byte {
	return evilHash()
}
