package boosted

import (
	"math/rand"
	"net/rpc"
)

type PirRpcServer struct {
	*pirServerPunc
	randSource *rand.Rand
	db         []Row
	server     *rpc.Server
}

func NewPirRpcServer(db []Row) *PirRpcServer {
	randSource := RandSource()
	driver := PirRpcServer{
		pirServerPunc: NewPirServerPunc(randSource, db),
		randSource:    randSource,
		db:            db,
	}
	return &driver
}

func (driver *PirRpcServer) SetDBDimensions(dim DBDimensions, none *int) error {
	driver.db = MakeDBWithDimensions(dim)
	driver.pirServerPunc = NewPirServerPunc(driver.randSource, driver.db)

	return nil
}

func (driver *PirRpcServer) SetRecordValue(rec RecordIndexVal, none *int) error {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	driver.pirServerPunc = NewPirServerPunc(driver.randSource, driver.db)
	return nil
}
