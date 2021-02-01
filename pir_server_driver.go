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
	PirUpdatableServer
	DB

	Configure(config TestConfig, none *int) error

	AddRows(numRows int, none *int) error
	DeleteRows(numRows int, none *int) error

	StartCpuProfile(int, *int) error
	StopCpuProfile(none int, out *string) error
	ResetMetrics(none int, none2 *int) error
	GetOfflineTimer(none int, out *time.Duration) error
	GetOnlineTimer(none int, out *time.Duration) error
	GetOfflineBytes(none int, out *int) error
	GetOnlineBytes(none int, out *int) error
}

type pirServerDriver struct {
	PirDB
	server *pirServerUpdatable

	config TestConfig

	randSource *rand.Rand
	updatable  bool

	profBuf bytes.Buffer

	// For profiling
	hintTime, answerTime      time.Duration
	offlineBytes, onlineBytes int
}

func registerExtraTypes() {
	gob.Register(&PuncturedSet{})
	gob.Register(&puncturedGGMSet{})
}

func NewPirServerDriver() (*pirServerDriver, error) {
	randSource := RandSource()
	registerExtraTypes()
	server := NewPirServerUpdatable(randSource)
	driver := pirServerDriver{
		PirDB:      server,
		server:     server,
		randSource: randSource,
	}
	return &driver, nil
}

func (driver *pirServerDriver) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	reqSize, err := SerializedSizeOf(req)
	if err != nil {
		return err
	}
	driver.offlineBytes += reqSize

	start := time.Now()
	if err = driver.server.KeyUpdates(req, resp); err != nil {
		return err
	}
	driver.hintTime += time.Since(start)

	respSize, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	driver.offlineBytes += respSize
	return nil
}

func (driver *pirServerDriver) Hint(req HintReq, resp *HintResp) error {
	reqSize, err := SerializedSizeOf(req)
	if err != nil {
		return err
	}
	driver.offlineBytes += reqSize

	start := time.Now()
	if err = driver.PirDB.Hint(req, resp); err != nil {
		return err
	}
	driver.hintTime += time.Since(start)

	respSize, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	driver.offlineBytes += respSize
	return nil
}

func (driver *pirServerDriver) Answer(q QueryReq, resp *QueryResp) error {
	reqSize, err := SerializedSizeOf(q)
	if err != nil {
		return err
	}
	driver.onlineBytes += reqSize

	start := time.Now()
	err = driver.PirDB.Answer(q, resp)
	driver.answerTime += time.Since(start)

	respSize, err := SerializedSizeOf(resp)
	if err != nil {
		return err
	}
	driver.onlineBytes += respSize
	return nil
}

func (driver *pirServerDriver) Configure(config TestConfig, none *int) (err error) {
	driver.config = config
	driver.updatable = config.Updatable
	if config.RandSeed > 0 {
		driver.randSource = rand.New(rand.NewSource(config.RandSeed))
	}

	db := MakeDB(driver.randSource, config.NumRows, config.RowLen)
	keys := MakeKeys(driver.randSource, config.NumRows)
	for _, preset := range config.PresetRows {
		copy(db[preset.Index], preset.Value)
		keys[preset.Index] = preset.Key
	}

	if config.Updatable {
		driver.server = NewPirServerUpdatable(driver.randSource)
		driver.server.AddRows(keys, db)
		driver.PirDB = driver.server
	} else {
		driver.PirDB = NewPirServerByType(config.PirType, driver.randSource, flattenDb(db), len(db), len(db[0]))
		driver.server = nil
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
	driver.server.AddRows(newKeys, newVals)
	return nil
}

func (driver *pirServerDriver) DeleteRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot DeleteRows from Non-Updatable PIR server")
	}
	keys := driver.server.SomeKeys(numRows)
	driver.server.DeleteRows(keys)
	return nil
}

func (driver *pirServerDriver) GetRow(idx int, row *RowIndexVal) error {
	return driver.PirDB.GetRow(idx, row)
}

func (driver *pirServerDriver) NumRows(none int, out *int) error {
	return driver.PirDB.NumRows(none, out)
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

func (driver *pirServerDriver) GetOfflineTimer(none int, out *time.Duration) error {
	*out = driver.hintTime
	return nil
}

func (driver *pirServerDriver) GetOnlineTimer(none int, out *time.Duration) error {
	*out = driver.answerTime
	return nil
}

func (driver *pirServerDriver) GetOfflineBytes(none int, out *int) error {
	*out = driver.offlineBytes
	return nil
}

func (driver *pirServerDriver) GetOnlineBytes(none int, out *int) error {
	*out = driver.onlineBytes
	return nil
}

func (driver *pirServerDriver) ResetMetrics(none int, none2 *int) error {
	driver.hintTime = 0
	driver.answerTime = 0
	driver.offlineBytes = 0
	driver.onlineBytes = 0
	return nil
}
