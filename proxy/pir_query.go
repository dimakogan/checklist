package main

import (
  "bytes"
  "crypto/sha256"
  "log"
)

var evilURLs = []string {
  // We put a bunch of bogus URLs here to make sure
  // that the delta values are all at most 31 bits long.
  "a", "b", "c", "d", "e", "f", "g", "h", "i", "j",

  // Here are the real URLs
  "testsafebrowsing.appspot.com/s/phishing.html",
  "testsafebrowsing.appspot.com/s/unwanted.html",
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

  for _, v := range evilURLs {
    h := computeHash([]byte(v))
    if bytes.Equal(h[0:PartialHashLen], hashIn) {
      return h
    }
  }

  return []byte{}
}

