package boosted

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"runtime/pprof"
)

type PirRpcServer struct {
	PirServer
	randSource *rand.Rand
	db         []Row
	keys       []uint32
	pirType    string
	serverDB   PirDB

	profBuf bytes.Buffer
}

func registerExtraTypes() {
	gob.Register(&ShiftedSet{})
	gob.Register(&puncturedGGMSet{})
}

func NewPirRpcServer(keys []uint32, db []Row) (*PirRpcServer, error) {
	randSource := RandSource()
	registerExtraTypes()
	server := NewPirServerUpdatable(randSource, false, keys, db)
	driver := PirRpcServer{
		PirServer:  server,
		randSource: randSource,
		pirType:    "punc",
		db:         db,
		keys:       keys,
	}
	return &driver, nil
}

func (driver *PirRpcServer) ResetDBDimensions(dim DBDimensions, none *int) (err error) {
	driver.db = MakeDBWithDimensions(dim)
	driver.keys = MakeKeys(dim.NumRecords)
	driver.reloadServer()
	return nil

}

func (driver *PirRpcServer) ChangeDBDimensions(delta int, none *int) (err error) {
	if delta > 0 {
		newVals := MakeDBWithDimensions(DBDimensions{delta, len(driver.db[0])})
		newKeys := MakeKeys(delta)

		driver.serverDB.AddRows(newKeys, newVals)
	}

	return nil

}

func (driver *PirRpcServer) SetRecordValue(rec RecordIndexVal, none *int) (err error) {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	driver.keys[rec.Index] = rec.Key
	driver.reloadServer()
	return nil
}

func (driver *PirRpcServer) SetPIRType(pirType string, none *int) error {
	driver.pirType = pirType
	driver.reloadServer()
	return nil
}

func (driver *PirRpcServer) StartCpuProfile(int, *int) error {
	driver.profBuf.Reset()
	return pprof.StartCPUProfile(&driver.profBuf)
}

func (driver *PirRpcServer) StopCpuProfile(none int, out *string) error {
	pprof.StopCPUProfile()
	*out = driver.profBuf.String()
	return nil
}

func (driver *PirRpcServer) reloadServer() {
	switch driver.pirType {
	case "punc":
		driver.PirServer = NewPirServerUpdatable(driver.randSource, false, driver.keys, driver.db)
	case "matrix":
		driver.PirServer = NewPirServerUpdatable(driver.randSource, true, driver.keys, driver.db)
	}
}
