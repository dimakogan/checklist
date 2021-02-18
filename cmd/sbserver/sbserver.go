package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"

	b "github.com/dimakogan/boosted-pir"
	sb "github.com/dimakogan/boosted-pir/safebrowsing"
	"github.com/rocketlaunchr/https-go"
	"github.com/ugorji/go/codec"
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
	useTLS := false
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

	var conn io.Closer
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		conn.Close()
	}()

	server := rpc.NewServer()
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}
	if useTLS {
		// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
		server.HandleHTTP("/", "/debug")

		log.Printf("Serving RPC server over HTTPS on port %d\n", *port)
		// Use self-signed certificate
		httpServer, _ := https.Server(fmt.Sprintf("%d", *port), https.GenerateOptions{Host: "checklist.app"})
		conn = httpServer
		err = httpServer.ListenAndServeTLS("", "")
		if err == http.ErrServerClosed {
			log.Println("Server shutdown")
		} else if err != nil {
			log.Fatal("Failed to http.Serve, %w", err)
		}
	} else {
		serveTCP(server, *port)
	}
}

func serveTCP(server *rpc.Server, port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen tcp: %v", err)
	}
	log.Printf("Serving RPC server over TCP on port %d\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("TCP Accept failed: %+v\n", err)
		}
		go server.ServeCodec(codec.GoRpc.ServerCodec(conn, b.CodecHandle()))
	}
}
