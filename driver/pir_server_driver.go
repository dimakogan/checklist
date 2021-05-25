package driver

import (
	"fmt"
	"math/rand"
	"time"

	"checklist/pir"
	"checklist/updatable"
)

type PirServerDriver interface {
	updatable.UpdatableServer

	Configure(config TestConfig, none *int) error

	AddRows(numRows int, none *int) error
	DeleteRows(numRows int, none *int) error
	GetRow(idx int, row *RowIndexVal) error
	NumRows(none int, out *int) error
	NumKeys(none int, out *int) error
	RowLen(none int, out *int) error

	ResetMetrics(none int, none2 *int) error
	GetOfflineTimer(none int, out *time.Duration) error
	GetOnlineTimer(none int, out *time.Duration) error
	GetOfflineBytes(none int, out *int) error
	GetOnlineBytes(none int, out *int) error
}

type RowIndexVal struct {
	Index int
	Key   uint32
	Value pir.Row
}

type serverDriver struct {
	updatableServer *updatable.Server
	staticDB        *pir.StaticDB

	config TestConfig

	randSource *rand.Rand
	updatable  bool

	// For profiling
	hintTime, answerTime      time.Duration
	offlineBytes, onlineBytes int
}

func NewServerDriver() (*serverDriver, error) {
	randSource := pir.RandSource()
	db := updatable.NewUpdatableServer()
	driver := serverDriver{
		updatableServer: db,
		staticDB:        &db.StaticDB,
		randSource:      randSource,
	}
	return &driver, nil
}

func (driver *serverDriver) KeyUpdates(req updatable.KeyUpdatesReq, resp *updatable.KeyUpdatesResp) error {
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

func (driver *serverDriver) Hint(req pir.HintReq, resp *pir.HintResp) (err error) {
	if driver.config.MeasureBandwidth {
		reqSize, err := SerializedSizeOf(req)
		if err != nil {
			return err
		}
		driver.offlineBytes += reqSize
	}

	start := time.Now()
	if *resp, err = req.Process(*driver.staticDB); err != nil {
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

func (driver *serverDriver) Answer(q pir.QueryReq, resp *interface{}) (err error) {
	if driver.config.MeasureBandwidth {
		reqSize, err := SerializedSizeOf(q)
		if err != nil {
			return err
		}
		driver.onlineBytes += reqSize
	}

	start := time.Now()
	if *resp, err = q.Process(*driver.staticDB); err != nil {
		return err
	}
	driver.answerTime += time.Since(start)

	if driver.config.MeasureBandwidth {
		respSize, _ := SerializedSizeOf(resp)
		driver.onlineBytes += respSize
	}
	return nil
}

func (driver *serverDriver) Configure(config TestConfig, none *int) (err error) {
	driver.config = config
	driver.updatable = config.Updatable
	if config.DataRandSeed > 0 {
		driver.randSource = rand.New(rand.NewSource(config.DataRandSeed))
	}

	rows := pir.MakeRows(driver.randSource, config.NumRows, config.RowLen)
	keys := pir.MakeKeys(driver.randSource, config.NumRows)
	for _, preset := range config.PresetRows {
		copy(rows[preset.Index], preset.Value)
		keys[preset.Index] = preset.Key
	}

	if config.Updatable {
		driver.updatableServer = updatable.NewUpdatableServer()
		driver.updatableServer.AddRows(keys, rows)
		driver.staticDB = &driver.updatableServer.StaticDB
	} else {
		driver.staticDB = pir.StaticDBFromRows(rows)
		driver.updatableServer = nil
	}

	driver.ResetMetrics(0, nil)
	return nil

}

func (driver *serverDriver) AddRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot AddRows to Non-Updatable PIR server")
	}
	newVals := pir.MakeRows(driver.randSource, numRows, driver.config.RowLen)
	newKeys := pir.MakeKeys(driver.randSource, numRows)
	// newKeys := make([]uint32, len(newVals))
	// curNumRows := driver.updatableServer.NumRows
	// for i := 0; i < len(newVals); i++ {
	// 	newKeys[i] = uint32(curNumRows + i)
	// }
	driver.updatableServer.AddRows(newKeys, newVals)
	return nil
}

func (driver *serverDriver) DeleteRows(numRows int, none *int) (err error) {
	if !driver.updatable {
		return fmt.Errorf("Cannot DeleteRows from Non-Updatable PIR server")
	}
	keys := driver.updatableServer.SomeKeys(numRows)
	driver.updatableServer.DeleteRows(keys)
	return nil
}

func (driver *serverDriver) GetRow(idx int, row *RowIndexVal) error {
	row.Index = idx
	var err error
	if driver.updatableServer != nil {
		row.Key, row.Value, err = driver.updatableServer.Row(idx)
		return err
	}
	row.Value = driver.staticDB.Row(idx)
	return nil
}

func (driver *serverDriver) NumRows(none int, out *int) error {
	*out = driver.staticDB.NumRows
	return nil
}

func (driver *serverDriver) NumKeys(none int, out *int) error {
	*out = driver.updatableServer.NumKeys()
	return nil
}

func (driver *serverDriver) RowLen(none int, out *int) error {
	*out = driver.staticDB.RowLen
	return nil
}

func (driver *serverDriver) GetOfflineTimer(none int, out *time.Duration) error {
	*out = driver.hintTime
	return nil
}

func (driver *serverDriver) GetOnlineTimer(none int, out *time.Duration) error {
	*out = driver.answerTime
	return nil
}

func (driver *serverDriver) GetOfflineBytes(none int, out *int) error {
	*out = driver.offlineBytes
	return nil
}

func (driver *serverDriver) GetOnlineBytes(none int, out *int) error {
	*out = driver.onlineBytes
	return nil
}

func (driver *serverDriver) ResetMetrics(none int, none2 *int) error {
	driver.hintTime = 0
	driver.answerTime = 0
	driver.offlineBytes = 0
	driver.onlineBytes = 0
	return nil
}
