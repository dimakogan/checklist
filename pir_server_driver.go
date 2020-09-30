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
	GetRow(idx int, row *RowIndexVal) error
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
	server := NewPirServerUpdatable(driver.randSource, driver.pirType)
	db := MakeDB(driver.randSource, config.NumRows, config.RowLen)
	keys := MakeKeys(driver.randSource, config.NumRows)
	for _, preset := range config.PresetRows {
		copy(db[preset.Index], preset.Value)
		keys[preset.Index] = preset.Key
	}
	server.AddRows(keys, db)

	driver.ResetTimers(0, nil)
	driver.config = config
	driver.pirType = config.PirType
	driver.PirServer = server
	driver.PirDB = server
	return nil

}

func (driver *pirServerDriver) AddRows(numRows int, none *int) (err error) {
	newVals := MakeDB(driver.randSource, numRows, driver.config.RowLen)
	newKeys := MakeKeys(driver.randSource, numRows)
	driver.PirDB.AddRows(newKeys, newVals)
	return nil
}

func (driver *pirServerDriver) DeleteRows(numRows int, none *int) (err error) {
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

func (driver *pirServerDriver) GetRow(idx int, row *RowIndexVal) error {
	keys, rows := driver.Elements(idx, idx+1)
	if keys == nil {
		return fmt.Errorf("Index %d out of bounds", idx)
	}
	if len(keys) != 1 || len(rows) != 1 {
		panic(fmt.Sprintf("Invalid returned slice length: %d, %d", len(keys), len(rows)))
	}
	row.Index = idx
	row.Key = keys[0]
	row.Value = rows[0]

	return nil
}
