package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"os"
	"time"

	"github.com/google/safebrowsing"
)

// From google/safebrowsing/hash.go

// hashPrefix represents a SHA256 hash. It may either be
// be full, where len(Hash) == maxHashPrefixLength, or
// be partial, where len(Hash) >= minHashPrefixLength.
type hashPrefix string
type hashPrefixes []hashPrefix

// From google/safebrowsing/database.go
type threatsForUpdate map[safebrowsing.ThreatDescriptor]partialHashes
type partialHashes struct {
	// Since the Hashes field is only needed when storing to disk and when
	// updating, this field is cleared except for when it is in use.
	// This is done to reduce memory usage as the contents of this can be
	// regenerated from the tfl.
	Hashes hashPrefixes

	SHA256 []byte // The SHA256 over Hashes
	State  []byte // Arbitrary binary blob to synchronize state with API
}

// databaseFormat is a light struct used only for gob encoding and decoding.
// As written to disk, the format of the database file is basically the gzip
// compressed version of the gob encoding of databaseFormat.
type databaseFormat struct {
	Table threatsForUpdate
	Time  time.Time
}

func (p hashPrefixes) SHA256() []byte {
	hash := sha256.New()
	for _, b := range p {
		hash.Write([]byte(b))
	}
	return hash.Sum(nil)
}

// loadDatabase loads the database state from a file.
func loadDatabase(path string) (db databaseFormat, err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return db, err
	}
	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return db, err
	}
	defer func() {
		if zerr := gz.Close(); err == nil {
			err = zerr
		}
	}()

	decoder := gob.NewDecoder(gz)
	if err = decoder.Decode(&db); err != nil {
		return db, err
	}
	for _, dv := range db.Table {
		if !bytes.Equal(dv.SHA256, dv.Hashes.SHA256()) {
			return db, errors.New("safebrowsing: threat list SHA256 mismatch")
		}
	}
	return db, nil
}
