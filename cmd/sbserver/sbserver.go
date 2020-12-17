package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	port := flag.Int("p", 12345, "Listening port")
	blockList := flag.String("f", "", "URL block list file")
	b.InitTestFlags()
	flag.Parse()

	configs := b.TestConfigs()
	if len(configs) != 1 {
		log.Fatalf("Too many config options: %v\n", configs)
	}
	config := configs[0]

	if len(*blockList) != 0 {
		readBlockedURLs(*blockList, &config)
	}

	config.NumRows = len(config.PresetRows)

	driver, err := b.NewPirServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	err = driver.Configure(config, nil)
	if err != nil {
		log.Fatalf("Failed to configure server: %s\n", err)
	}

	server := rpc.NewServer()
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", *port)}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		httpServer.Close()
	}()

	log.Printf("Serving RPC server on port port %d\n", *port)
	// Start accept incoming HTTP connections
	e := httpServer.ListenAndServe()
	if e == http.ErrServerClosed {
		log.Println("Server shutdown")
	} else if e != nil {
		log.Fatal("Failed to http.Serve, %w", e)
	}
}
