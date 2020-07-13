package main

import (
	"crypto/sha256"
)

const evilHost = "testsafebrowsing.appspot.com/s/phishing.html"

// This is the function that will make the PIR query.
// It takes as input an index `idx` and outputs the
// full SHA256 hash of the URL corresponding to that index.
func queryForHash(idx int) []byte {
	hash := sha256.New()
	hash.Write([]byte(evilHost))
	return hash.Sum(nil)
}
