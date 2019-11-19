package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"

	b "github.com/dimakogan/boosted-pir"
)

type ServerTestDriver struct {
	b.PIRServer
	randSource *rand.Rand
	db         []b.Row
	server     *rpc.Server
}

func NewServerTestDriver(db []b.Row) (*ServerTestDriver, error) {
	randSource := b.RandSource()
	driver := ServerTestDriver{
		PIRServer:  b.NewPirServerPunc(randSource, db),
		randSource: randSource,
		db:         db,
		server:     rpc.NewServer(),
	}
	if err := driver.server.RegisterName("PIRServer", &driver); err != nil {
		return nil, fmt.Errorf("Failed to register PIRServer, %w", err)
	}

	// registers an HTTP handler for RPC messages on rpcPath, and a debugging handler on debugPath
	driver.server.HandleHTTP("/", "/debug")

	return &driver, nil
}

func (driver *ServerTestDriver) serve(addr string) error {
	// Listen to TPC connections on port 1234
	listener, e := net.Listen("tcp", addr)
	if e != nil {
		return fmt.Errorf("Listen error, %w ", e)
	}
	log.Printf("Serving RPC server on %s", addr)
	// Start accept incoming HTTP connections
	return http.Serve(listener, nil)
}

func (driver *ServerTestDriver) SetDBDimensions(dim b.DBDimensions, none *int) error {
	driver.db = b.MakeDBWithDimensions(dim)
	driver.PIRServer = b.NewPirServerPunc(driver.randSource, driver.db)

	return nil
}

func (driver *ServerTestDriver) SetRecordValue(rec b.RecordIndexVal, none *int) error {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	return nil
}

func main() {
	// Some easy to test initial values.
	var db = []b.Row{{'A'}, {'B'}, {'C'}, {'D'}}

	driver, err := NewServerTestDriver(db)
	if err != nil {
		log.Fatal("Failed to initialize driver, ", err)
	}

	if err = driver.serve(":1234"); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
}
