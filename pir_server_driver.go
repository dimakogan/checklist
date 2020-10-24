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

	Configure(config TestConfig, none *int) error
	AddRows(numRows int, none *int) error
	DeleteRows(numRows int, none *int) error
	StartCpuProfile(int, *int) error
	StopCpuProfile(none int, out *string) error
	ResetTimers(none int, none2 *int) error
	GetHintTimer(none int, out *time.Duration) error
	GetAnswerTimer(none int, out *time.Duration) error
}

type pirServerDriver struct {
	PirServer
	PirDB

	config TestConfig

	randSource *rand.Rand
	pirType    PirType
	updatable  bool

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
	server := NewPirServerUpdatable(randSource, Punc)
	driver := pirServerDriver{
		PirServer:  server,
		PirDB:      server,
		randSource: randSource,
		pirType:    Punc,
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

func (driver *pirServerDriver) Configure(config TestConfig, none *int) (err error) {
	db := MakeDB(driver.randSource, config.NumRows, config.RowLen)
	keys := MakeKeys(driver.randSource, config.NumRows)
	for _, preset := range config.PresetRows {
		copy(db[preset.Index], preset.Value)
		keys[preset.Index] = preset.Key
	}

	if config.Updatable {
		server := NewPirServerUpdatable(driver.randSource, driver.pirType)
		server.AddRows(keys, db)
		driver.PirServer = server
		driver.PirDB = server
	} else {
		driver.PirServer = NewPirServerByType(config.PirType, driver.randSource, db)
		driver.PirDB = nil
	}

	driver.ResetTimers(0, nil)
	driver.config = config
	driver.pirType = config.PirType
	driver.updatable = config.Updatable
	return nil

}

func (driver *pirServerDriver) AddRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot AddRows to Non-Updatable PIR server")
	}
	newVals := MakeDB(driver.randSource, numRows, driver.config.RowLen)
	newKeys := MakeKeys(driver.randSource, numRows)
	driver.PirDB.AddRows(newKeys, newVals)
	return nil
}

func (driver *pirServerDriver) DeleteRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot DeleteRows from Non-Updatable PIR server")
	}
	keys := driver.PirDB.SomeKeys(numRows)
	driver.PirDB.DeleteRows(keys)
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
