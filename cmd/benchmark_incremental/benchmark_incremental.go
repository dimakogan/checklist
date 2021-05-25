package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	. "checklist/driver"
	"checklist/pir"
	"checklist/updatable"

	"gotest.tools/assert"
)

func main() {
	var ep ErrorPrinter

	config := new(Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()

	prof := NewProfiler(config.CpuProfile)
	defer prof.Close()

	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	fmt.Printf("%10s%12s%22s%22s%22s%15s%22s%22s%15s\n",
		"numRows",
		"updateSize",
		"UpdateServerTime[us]", "UpdateClientTime[us]", "UpdateBytesPerChange", "ClientBytes",
		"OnlineServerTime[us]", "OnlineClientTime[us]", "OnlineBytes")

	driver, err := config.ServerDriver()
	if err != nil {
		log.Fatalf("Failed to create driver: %s\n", err)
	}

	rand := pir.RandSource()

	var none int
	if err := driver.Configure(config.TestConfig, &none); err != nil {
		log.Fatalf("Failed to configure driver: %s\n", err)
	}

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driver, driver})

	err = client.Init()
	assert.NilError(ep, err)

	driver.ResetMetrics(0, &none)
	var clientUpdateTime, clientReadTime time.Duration

	var numBatches int
	if config.NumUpdates == 0 {
		numBatches = (config.NumRows-1)/(config.UpdateSize) + 1
	} else {
		numBatches = config.NumUpdates
	}

	for i := 0; i < numBatches; i++ {
		assert.NilError(ep, driver.AddRows(config.UpdateSize/2, &none))
		assert.NilError(ep, driver.DeleteRows(config.UpdateSize/2, &none))

		start := time.Now()
		client.Update()
		clientUpdateTime += time.Since(start)

		var rowIV RowIndexVal
		var numKeys int
		assert.NilError(ep, driver.NumKeys(none, &numKeys))
		assert.NilError(ep, driver.GetRow(rand.Intn(numKeys), &rowIV))

		start = time.Now()
		row, err := client.Read(rowIV.Key)
		clientReadTime += time.Since(start)
		assert.NilError(ep, err)
		assert.DeepEqual(ep, row, rowIV.Value)

		if config.Progress {
			fmt.Fprintf(os.Stderr, "%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, numBatches)
		}

		if i == numBatches-2 {
			runtime.GC()
			if memProf, err := os.Create("mem.prof"); err != nil {
				panic(err)
			} else {
				pprof.WriteHeapProfile(memProf)
				memProf.Close()
			}
		}
	}
	var serverOfflineTime, serverOnlineTime time.Duration
	var offlineBytes, onlineBytes int
	assert.NilError(ep, driver.GetOfflineTimer(0, &serverOfflineTime))
	assert.NilError(ep, driver.GetOnlineTimer(0, &serverOnlineTime))
	assert.NilError(ep, driver.GetOnlineBytes(0, &onlineBytes))
	assert.NilError(ep, driver.GetOfflineBytes(0, &offlineBytes))

	fmt.Printf("%10d%12d%22d%22d%22d%15d%22d%22d%15d\n",
		config.NumRows,
		config.UpdateSize,
		serverOfflineTime.Microseconds()/int64(numBatches),
		(clientUpdateTime-serverOfflineTime).Microseconds()/int64(numBatches),
		offlineBytes/(numBatches*config.UpdateSize),
		client.StorageNumBytes(SerializedSizeOf),
		serverOnlineTime.Microseconds()/int64(numBatches),
		(clientReadTime-serverOnlineTime).Microseconds()/int64(numBatches),
		onlineBytes/numBatches)
}
