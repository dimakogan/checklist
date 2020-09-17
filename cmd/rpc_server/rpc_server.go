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

	b "github.com/dimakogan/boosted-pir"
)

func main() {
	port := flag.Int("p", 12345, "Listening port")
	flag.Parse()

	// Some easy to test initial values.
	db := make([]b.Row, b.DEFAULT_CHUNK_SIZE)
	keys := make([]uint32, len(db))
	for i := 0; i < len(db); i++ {
		keys[i] = uint32(1000000 + i)
		db[i] = b.Row{byte('A' + i), byte('A' + i), byte('A' + i)}
	}

	driver, err := b.NewPirRpcServer(keys, db)
	if err != nil {
		log.Fatalf("Failed to create server: %s", err)
	}

	server := rpc.NewServer()
	if err := server.Register(driver); err != nil {
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
