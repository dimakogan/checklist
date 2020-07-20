package boosted

import (
	"math/rand"
	"net/rpc"
)

type PirRpcServer struct {
	PirServer
	randSource *rand.Rand
	db         []Row
	server     *rpc.Server
}

func NewPirRpcServer(db []Row) (*PirRpcServer, error) {
	randSource := RandSource()
	server, err := NewPirServerErasure(randSource, db, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return nil, err
	}
	driver := PirRpcServer{
		PirServer: server,
		randSource:    randSource,
		db:            db,
	}
	return &driver, nil
}

func (driver *PirRpcServer) SetDBDimensions(dim DBDimensions, none *int) (err error) {
	driver.db = MakeDBWithDimensions(dim)
	driver.PirServer, err = NewPirServerErasure(driver.randSource, driver.db, DEFAULT_CHUNK_SIZE)
	return err
}

func (driver *PirRpcServer) SetRecordValue(rec RecordIndexVal, none *int) (err error) {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	driver.PirServer, err = NewPirServerErasure(driver.randSource, driver.db, DEFAULT_CHUNK_SIZE)
	return err
}
