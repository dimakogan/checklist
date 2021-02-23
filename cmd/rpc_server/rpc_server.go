package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dimakogan/boosted-pir/rpc"

	b "github.com/dimakogan/boosted-pir"
	sb "github.com/dimakogan/boosted-pir/safebrowsing"
)

func readBlockedURLs(blockListFile string, config *b.TestConfig) {
	file, err := os.Open(blockListFile)
	if err != nil {
		log.Fatalf("Failed to open block list file %s: %s", blockListFile, err)
	}
	scanner := bufio.NewScanner(file)
	pos := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		partial, full := sb.ComputeHash([]byte(line))
		entry := b.RowIndexVal{
			Index: pos,
			Key:   binary.LittleEndian.Uint32(partial),
			Value: full,
		}
		pos++
		log.Printf("Evil URL hash prefix: %x, full: %x\n", entry.Key, entry.Value)
		config.PresetRows = append(config.PresetRows, entry) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}

func main() {
	config := b.NewConfig().WithServerFlags()
	blockList := config.FlagSet.String("f", "", "URL block list file")
	config.Parse()

	if len(*blockList) != 0 {
		readBlockedURLs(*blockList, &config.TestConfig)
	}

	config.NumRows = len(config.PresetRows)

	driver, err := b.NewPirServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	server, err := rpc.NewServer(config.Port, config.UseTLS)
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

	prof := b.NewProfiler(config.CpuProfile)
	defer prof.Close()

	err = server.Serve()
	if err != nil && !inShutdown {
		log.Fatalf("Failed to serve: %s", err)
	} else {
		fmt.Printf("Shutting down")
	}
}
