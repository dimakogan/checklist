package main

import (
  "crypto/sha256"
  "log"
)

var evilURLs = []string {
  "testsafebrowsing.appspot.com/s/phishing.html",
  //"testsafebrowsing.appspot.com/s/unwanted.html",
}

const PartialHashLen = 4
type PartialHash = []byte

func computeHash(url []byte) []byte {
	hash := sha256.New()
	hash.Write(url)
  return hash.Sum(nil)
}


// This function queries the PIR client for the
// list of 4-byte partial hashes.
func fetchPartialHashes() []PartialHash {
  out := make([]PartialHash, len(evilURLs))
  for i,v := range(evilURLs) {
    out[i] = computeHash([]byte(v))[0:PartialHashLen]
  }

  return out
}

// This is the function that will make the PIR query.
// It takes as input a hash prefix and outputs the
// full SHA256 hash of the URL corresponding to that index.
func queryForHash(hashIn []byte) []byte {
  log.Printf("Looking for hash = %v", hashIn)
	return computeHash([]byte(evilURLs[0]))
}

