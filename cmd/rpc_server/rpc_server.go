package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/rocketlaunchr/https-go"

	b "github.com/dimakogan/boosted-pir"
)

func main() {
	port := flag.Int("p", 12345, "Listening port")
	useTLS := flag.Bool("tls", true, "Should use TLS")
	flag.Parse()

	driver, err := b.NewPirServerDriver()
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	server := rpc.NewServer()
	if err := server.RegisterName("PirServerDriver", driver); err != nil {
		log.Fatalf("Failed to register PIRServer, %s", err)
	}

	// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	server.HandleHTTP("/", "/debug")

	log.Printf("Serving RPC server on port %d\n", *port)
	var httpServer *http.Server
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		httpServer.Close()
	}()

	if *useTLS {
		// Use self-signed certificate
		httpServer, _ = https.Server(fmt.Sprintf("%d", *port), https.GenerateOptions{Host: "checklist.app"})
		err = httpServer.ListenAndServeTLS("", "")
	} else {
		httpServer = &http.Server{Addr: fmt.Sprintf(":%d", *port)}
		err = httpServer.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		log.Println("Server shutdown")
	} else if err != nil {
		log.Fatal("Failed to http.Serve, %w", err)
	}
}
