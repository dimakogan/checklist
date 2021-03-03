package boosted

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"time"
)

type PirServerDriver interface {
	PirUpdatableServer

	Configure(config TestConfig, none *int) error

	AddRows(numRows int, none *int) error
	DeleteRows(numRows int, none *int) error
	GetRow(idx int, row *RowIndexVal) error
	NumRows(none int, out *int) error
	RowLen(none int, out *int) error

	ResetMetrics(none int, none2 *int) error
	GetOfflineTimer(none int, out *time.Duration) error
	GetOnlineTimer(none int, out *time.Duration) error
	GetOfflineBytes(none int, out *int) error
	GetOnlineBytes(none int, out *int) error
}

type pirServerDriver struct {
	server          PirServer
	updatableServer *pirUpdatableServer
	staticDB        *staticDB

	config TestConfig

	randSource *rand.Rand
	updatable  bool

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
	db := NewPirUpdatableServer()
	driver := pirServerDriver{
		updatableServer: db,
		staticDB:        &db.staticDB,
		randSource:      randSource,
	}
	return &driver, nil
}

func (driver *pirServerDriver) KeyUpdates(req KeyUpdatesReq, resp *KeyUpdatesResp) error {
	if driver.config.MeasureBandwidth {
		reqSize, err := SerializedSizeOf(req)
		if err != nil {
			return err
		}
		driver.offlineBytes += reqSize
	}

	start := time.Now()
	if err := driver.updatableServer.KeyUpdates(req, resp); err != nil {
		return err
	}
	driver.hintTime += time.Since(start)

	if driver.config.MeasureBandwidth {
		respSize, err := SerializedSizeOf(resp)
		if err != nil {
			return err
		}
		driver.offlineBytes += respSize
	}
	return nil
}

func (driver *pirServerDriver) Hint(req HintReq, resp *HintResp) error {
	if driver.config.MeasureBandwidth {
		reqSize, err := SerializedSizeOf(req)
		if err != nil {
			return err
		}
		driver.offlineBytes += reqSize
	}

	start := time.Now()
	if err := driver.server.Hint(req, resp); err != nil {
		return err
	}
	driver.hintTime += time.Since(start)

	if driver.config.MeasureBandwidth {
		respSize, err := SerializedSizeOf(resp)
		if err != nil {
			return err
		}
		driver.offlineBytes += respSize
	}
	return nil
}

func (driver *pirServerDriver) Answer(q QueryReq, resp *QueryResp) error {
	if driver.config.MeasureBandwidth {
		reqSize, err := SerializedSizeOf(q)
		if err != nil {
			return err
		}
		driver.onlineBytes += reqSize
	}

	start := time.Now()
	if err := driver.server.Answer(q, resp); err != nil {
		return err
	}
	driver.answerTime += time.Since(start)

	if driver.config.MeasureBandwidth {
		respSize, _ := SerializedSizeOf(resp)
		driver.onlineBytes += respSize
	}
	return nil
}

func (driver *pirServerDriver) Configure(config TestConfig, none *int) (err error) {
	driver.config = config
	driver.updatable = config.Updatable
	if config.RandSeed > 0 {
		driver.randSource = rand.New(rand.NewSource(config.RandSeed))
	}

	rows := MakeRows(driver.randSource, config.NumRows, config.RowLen)
	keys := MakeKeys(driver.randSource, config.NumRows)
	for _, preset := range config.PresetRows {
		copy(rows[preset.Index], preset.Value)
		keys[preset.Index] = preset.Key
	}

	if config.Updatable {
		driver.updatableServer = NewPirUpdatableServer()
		driver.updatableServer.AddRows(keys, rows)
		driver.staticDB = &driver.updatableServer.staticDB
		driver.server = driver.updatableServer
	} else {
		driver.staticDB = &staticDB{config.NumRows, config.RowLen, flattenDb(rows)}
		driver.updatableServer = nil
		driver.server = NewPirServerByType(config.PirType, driver.staticDB)
	}

	driver.ResetMetrics(0, nil)
	return nil

}

func (driver *pirServerDriver) AddRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot AddRows to Non-Updatable PIR server")
	}
	newVals := MakeRows(driver.randSource, numRows, driver.config.RowLen)
	newKeys := MakeKeys(driver.randSource, numRows)
	driver.updatableServer.AddRows(newKeys, newVals)
	return nil
}

func (driver *pirServerDriver) DeleteRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot DeleteRows from Non-Updatable PIR server")
	}
	keys := driver.updatableServer.SomeKeys(numRows)
	driver.updatableServer.DeleteRows(keys)
	return nil
}

func (driver *pirServerDriver) GetRow(idx int, row *RowIndexVal) error {
	if driver.updatableServer != nil {
		return driver.updatableServer.GetRow(idx, row)
	} else {
		return driver.staticDB.GetRow(idx, row)
	}
}

func (driver *pirServerDriver) NumRows(none int, out *int) error {
	*out = driver.staticDB.numRows
	return nil
}

func (driver *pirServerDriver) RowLen(none int, out *int) error {
	*out = driver.staticDB.rowLen
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
