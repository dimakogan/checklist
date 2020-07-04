package main

import (
	"errors"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/google/safebrowsing"
)

type LocalIndex struct {
	databaseFile    string
	watcher         *fsnotify.Watcher
	hashPrefixIndex map[hashPrefix]int
	lock            sync.RWMutex
}

func NewLocalIndex(dbFile string) *LocalIndex {
	var err error
	ind := new(LocalIndex)
	ind.databaseFile = dbFile
	ind.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Cannot create watcher: %v", err)
	}

	if err := ind.watcher.Add(ind.databaseFile); err != nil {
		log.Fatalf("Cannot watch file: %v", err)
	}

	// Update hash map whenever database changes
	go func() {
		for {
			select {
			case event, ok := <-ind.watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					ind.update()
				}
			case err, ok := <-ind.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	return ind
}

func (ind *LocalIndex) Close() {
	ind.watcher.Close()
}

func (ind *LocalIndex) update() error {
	ind.lock.Lock()
	defer ind.lock.Unlock()

	db, err := loadDatabase(ind.databaseFile)
	if err != nil {
		log.Printf("Cannot load DB -- this may happen when update is in progress.")
		return err
	}

	ind.hashPrefixIndex = make(map[hashPrefix]int)

	i := 0
	for _, list := range safebrowsing.DefaultThreatLists {
		table, ok := db.Table[list]
		if !ok {
			log.Fatal("Something went wrong")
		}

		for j := range table.Hashes {
			ind.hashPrefixIndex[table.Hashes[j]] = i
			i++
		}
	}

	return err
}

func (ind *LocalIndex) GetIndex(hash hashPrefix) (int, error) {
	ind.lock.RLock()
	defer ind.lock.RUnlock()

	if ind.hashPrefixIndex == nil {
		return -1, errors.New("Must run Update() before GetIndex()")
	}

	v, ok := ind.hashPrefixIndex[hash]
	if !ok {
		return -1, errors.New("Unknown prefix")
	}

	return v, nil
}
