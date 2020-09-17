package boosted

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"net/rpc"
	"runtime/pprof"
)

type PirRpcServer struct {
	PirServer
	randSource *rand.Rand
	db         []Row
	keys       []uint32
	pirType    string
	server     *rpc.Server

	profBuf bytes.Buffer
}

func registerExtraTypes() {
	gob.Register(&ShiftedSet{})
	gob.Register(&puncturedGGMSet{})
}

func NewPirRpcServer(db []Row) (*PirRpcServer, error) {
	randSource := RandSource()
	registerExtraTypes()
	//server, err := NewPirServerErasure(randSource, db, DEFAULT_CHUNK_SIZE)
	// if err != nil {
	// 	return nil, err
	// }
	server := NewPirServerPunc(randSource, db)
	driver := PirRpcServer{
		PirServer:  server,
		randSource: randSource,
		pirType:    "punc",
		db:         db,
		keys:       MakeKeys(len(db)),
	}
	return &driver, nil
}

func (driver *PirRpcServer) SetDBDimensions(dim DBDimensions, none *int) (err error) {
	driver.db = MakeDBWithDimensions(dim)
	driver.keys = MakeKeys(dim.NumRecords)
	driver.reloadServer()
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
