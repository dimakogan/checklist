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
	ResetMetrics(none int, none2 *int) error
	GetHintTimer(none int, out *time.Duration) error
	GetAnswerTimer(none int, out *time.Duration) error
	GetHintBytes(none int, out *int) error
	GetAnswerBytes(none int, out *int) error
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
	hintTime, answerTime   time.Duration
	hintBytes, answerBytes int
}

func registerExtraTypes() {
	gob.Register(&PuncturedSet{})
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
	reqSize, err := SerializedSizeOf(req)
	if err != nil {
		return err
	}
	driver.hintBytes += reqSize

	start := time.Now()
	if err = driver.PirServer.Hint(req, resp); err != nil {
		return err
	}
	driver.hintTime += time.Since(start)

	respSize, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	driver.hintBytes += respSize
	return nil
}

func (driver *pirServerDriver) Answer(q QueryReq, resp *QueryResp) error {
	reqSize, err := SerializedSizeOf(q)
	if err != nil {
		return err
	}
	driver.answerBytes += reqSize

	start := time.Now()
	err = driver.PirServer.Answer(q, resp)
	driver.answerTime += time.Since(start)

	respSize, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	driver.answerBytes += respSize
	return nil
}

func (driver *pirServerDriver) Configure(config TestConfig, none *int) (err error) {
	driver.config = config
	driver.pirType = config.PirType
	driver.updatable = config.Updatable

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

	driver.ResetMetrics(0, nil)
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

func (driver *pirServerDriver) GetHintBytes(none int, out *int) error {
	*out = driver.hintBytes
	return nil
}

func (driver *pirServerDriver) GetAnswerBytes(none int, out *int) error {
	*out = driver.answerBytes
	return nil
}

func (driver *pirServerDriver) ResetMetrics(none int, none2 *int) error {
	driver.hintTime = 0
	driver.answerTime = 0
	driver.hintBytes = 0
	driver.answerBytes = 0
	return nil
}
