package boosted

import (
	"math/rand"
	"net/rpc"
)

type ServerTestDriver struct {
	*pirServerPunc
	randSource *rand.Rand
	db         []Row
	server     *rpc.Server
}

func NewServerTestDriver(db []Row) *ServerTestDriver {
	randSource := RandSource()
	driver := ServerTestDriver{
		pirServerPunc: NewPirServerPunc(randSource, db),
		randSource:    randSource,
		db:            db,
	}
	return &driver
}

func (driver *ServerTestDriver) SetDBDimensions(dim DBDimensions, none *int) error {
	driver.db = MakeDBWithDimensions(dim)
	driver.pirServerPunc = NewPirServerPunc(driver.randSource, driver.db)

	return nil
}

func (driver *ServerTestDriver) SetRecordValue(rec RecordIndexVal, none *int) error {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	return nil
}
