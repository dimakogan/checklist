package boosted

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"runtime/pprof"
	"time"
)

type PirServerDriver interface {
	PirServer

	SetPIRType(pirType PirType, none *int) error
	ResetDBDimensions(dim DBDimensions, none *int) error
	AddRows(numRows int, none *int) error
	DeleteRows(numRows int, none *int) error
	SetRecordValue(rec RecordIndexVal, none *int) error
	StartCpuProfile(int, *int) error
	StopCpuProfile(none int, out *string) error
	GetRecord(idx int, record *RecordIndexVal) error
	ResetTimers(none int, none2 *int) error
	GetHintTimer(none int, out *time.Duration) error
	GetAnswerTimer(none int, out *time.Duration) error
}

type pirServerDriver struct {
	PirServer
	PirDB

	randSource *rand.Rand
	db         []Row
	keys       []uint32
	pirType    PirType

	profBuf bytes.Buffer

	// For profiling
	hintTime   time.Duration
	answerTime time.Duration
}

func registerExtraTypes() {
	gob.Register(&ShiftedSet{})
	gob.Register(&puncturedGGMSet{})
}

func NewPirServerDriver() (*pirServerDriver, error) {
	randSource := RandSource()
	registerExtraTypes()
	server := NewPirServerUpdatable(randSource, PirPuncturable)
	driver := pirServerDriver{
		PirServer:  server,
		PirDB:      server,
		randSource: randSource,
		pirType:    PirPuncturable,
	}
	return &driver, nil
}

func (driver *pirServerDriver) Hint(req HintReq, resp *HintResp) error {
	start := time.Now()
	err := driver.PirServer.Hint(req, resp)
	driver.hintTime += time.Since(start)
	return err
}

func (driver *pirServerDriver) Answer(q QueryReq, resp *QueryResp) error {
	start := time.Now()
	err := driver.PirServer.Answer(q, resp)
	driver.answerTime += time.Since(start)
	return err
}

func (driver *pirServerDriver) ResetDBDimensions(dim DBDimensions, none *int) (err error) {
	driver.db = MakeDBWithDimensions(driver.randSource, dim)
	driver.keys = MakeKeys(driver.randSource, dim.NumRecords)
	driver.reloadServer()
	return nil

}

func (driver *pirServerDriver) AddRows(numRows int, none *int) (err error) {
	newVals := MakeDBWithDimensions(driver.randSource, DBDimensions{numRows, len(driver.db[0])})
	newKeys := MakeKeys(driver.randSource, numRows)
	driver.db = append(driver.db, newVals...)
	driver.keys = append(driver.keys, newKeys...)
	driver.PirDB.AddRows(newKeys, newVals)
	return nil
}

func (driver *pirServerDriver) DeleteRows(numRows int, none *int) (err error) {
	driver.PirDB.DeleteRows(driver.keys[0:numRows])
	driver.db = driver.db[numRows:]
	driver.keys = driver.keys[numRows:]
	return nil
}

func (driver *pirServerDriver) SetRecordValue(rec RecordIndexVal, none *int) (err error) {
	// There is a single shallow copy, so this should propagate into the PIR serve rinstance.
	driver.db[rec.Index] = rec.Value
	driver.keys[rec.Index] = rec.Key
	driver.reloadServer()
	return nil
}

func (driver *pirServerDriver) SetPIRType(pirType PirType, none *int) error {
	driver.pirType = pirType
	driver.reloadServer()
	return nil
}

func (driver *pirServerDriver) StartCpuProfile(int, *int) error {
	driver.profBuf.Reset()
	return pprof.StartCPUProfile(&driver.profBuf)
}

func (driver *pirServerDriver) StopCpuProfile(none int, out *string) error {
	pprof.StopCPUProfile()
	*out = driver.profBuf.String()
	return nil
}

func (driver *pirServerDriver) GetHintTimer(none int, out *time.Duration) error {
	*out = driver.hintTime
	return nil
}

func (driver *pirServerDriver) GetAnswerTimer(none int, out *time.Duration) error {
	*out = driver.answerTime
	return nil
}

func (driver *pirServerDriver) ResetTimers(none int, none2 *int) error {
	driver.hintTime = 0
	driver.answerTime = 0
	return nil
}

func (driver *pirServerDriver) reloadServer() {
	server := NewPirServerUpdatable(driver.randSource, driver.pirType)
	driver.PirServer = server
	driver.PirDB = server
	driver.PirDB.AddRows(driver.keys, driver.db)

	// Reset timers
	driver.hintTime = 0
	driver.answerTime = 0
}

func (driver *pirServerDriver) GetRecord(idx int, record *RecordIndexVal) error {
	if idx >= len(driver.db) {
		return fmt.Errorf("Index %d out of bounds %d", idx, len(driver.db))
	}
	record.Index = idx
	record.Key = driver.keys[idx]
	record.Value = driver.db[idx]
	return nil
}
