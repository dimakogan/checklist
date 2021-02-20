package main

import (
	"bufio"
	"encoding/binary"
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

	"github.com/rocketlaunchr/https-go"
	"github.com/ugorji/go/codec"

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

	server := rpc.NewServer()
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	prof := b.NewCPUProfiler(config.CpuProfile)
	defer prof.Close()

	if config.UseTLS {
		// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
		server.HandleHTTP("/", "/debug")

		log.Printf("Serving RPC server over HTTPS on port %d\n", config.Port)
		// Use self-signed certificate
		httpServer, _ := https.Server(fmt.Sprintf("%d", config.Port), https.GenerateOptions{Host: "checklist.app", ECDSACurve: "P256"})
		closeOnSignal(httpServer)
		err = httpServer.ListenAndServeTLS("", "")
		if err == http.ErrServerClosed {
			log.Println("Server shutdown")
		} else if err != nil {
			log.Fatal("Failed to http.Serve, %w", err)
		}
	} else {
		serveTCP(server, config.Port)
	}
}

func serveTCP(server *rpc.Server, port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen tcp: %v", err)
	}
	log.Printf("Serving RPC server over TCP on port %d\n", port)
	closeOnSignal(ln)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("TCP Accept failed: %+v\n", err)
			return
		}
		go server.ServeCodec(codec.GoRpc.ServerCodec(conn, b.CodecHandle()))
	}
}

func closeOnSignal(conn io.Closer) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		conn.Close()
	}()
}
