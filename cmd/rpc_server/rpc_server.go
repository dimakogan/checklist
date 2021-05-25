package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	. "checklist/driver"
	"checklist/rpc"

	sb "checklist/safebrowsing"
)

func readBlockedURLs(blockListFile string, config *TestConfig) {
	file, err := os.Open(blockListFile)
	if err != nil {
		log.Fatalf("Failed to open block list file %s: %s", blockListFile, err)
	}
	partial, full, err := sb.ReadBlockedURLs(file)
	if err != nil {
		log.Fatalf("Failed to read blocked urls from %s: %s", blockListFile, err)
	}
	if len(partial) != len(full) {
		log.Fatalf("Invalid number of partial %d and full %d", len(partial), len(full))
	}
	for i := range partial {
		entry := RowIndexVal{
			Index: i,
			Key:   partial[i],
			Value: full[i],
		}
		log.Printf("Evil URL hash prefix: %x, full: %x\n", entry.Key, entry.Value)
		config.PresetRows = append(config.PresetRows, entry) // Println will add back the final '\n'
	}
}

func main() {
	config := new(Config).AddPirFlags().AddServerFlags()
	blockList := config.FlagSet.String("f", "", "URL block list file")
	config.Parse()

	if len(*blockList) != 0 {
		readBlockedURLs(*blockList, &config.TestConfig)
		config.NumRows = len(config.PresetRows)
	}

	driver, err := NewServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}
	var none int
	err = driver.Configure(config.TestConfig, &none)
	if err != nil {
		log.Fatalf("Failed to configure server: %s\n", err)
	}

	server, err := rpc.NewServer(config.Port, config.UseTLS, RegisteredTypes())
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	var inShutdown bool
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		inShutdown = true
		server.Close()
	}()

	prof := NewProfiler(config.CpuProfile)
	defer prof.Close()

	err = server.Serve()
	if err != nil && !inShutdown {
		log.Fatalf("Failed to serve: %s", err)
	} else {
		fmt.Printf("Shutting down")
	}
}
