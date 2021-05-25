package main

import (
	"fmt"
	"log"
	"os"
	"path"
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
	fmt.Printf("%15s%22s%22s%22s%15s%22s%22s%15s\n",
		"NumUpdates",
		"UpdateServerTime[us]", "UpdateClientTime[us]", "UpdateBytes", "ClientBytes",
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
	driver.ResetMetrics(0, &none)

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driver, driver})

	var clientUpdateTime, clientReadTime time.Duration

	if config.NumUpdates == 0 {
		config.NumUpdates = (config.NumRows-1)/config.UpdateSize + 1
	}
	for i := 0; i < config.NumUpdates; i++ {
		assert.NilError(ep, driver.AddRows(config.UpdateSize/2, &none))
		assert.NilError(ep, driver.DeleteRows(config.UpdateSize/2, &none))

		driver.ResetMetrics(0, &none)

		start := time.Now()
		client.Update()
		clientUpdateTime = time.Since(start)

		var rowIV RowIndexVal
		var numKeys int
		assert.NilError(ep, driver.NumKeys(none, &numKeys))
		assert.NilError(ep, driver.GetRow(rand.Intn(numKeys), &rowIV))

		start = time.Now()
		row, err := client.Read(rowIV.Key)
		clientReadTime = time.Since(start)
		assert.NilError(ep, err)
		assert.DeepEqual(ep, row, rowIV.Value)

		var serverOfflineTime, serverOnlineTime time.Duration
		var offlineBytes, onlineBytes int
		assert.NilError(ep, driver.GetOfflineTimer(0, &serverOfflineTime))
		assert.NilError(ep, driver.GetOnlineTimer(0, &serverOnlineTime))
		assert.NilError(ep, driver.GetOnlineBytes(0, &onlineBytes))
		assert.NilError(ep, driver.GetOfflineBytes(0, &offlineBytes))

		fmt.Printf("%15d%22d%22d%22d%15d%22d%22d%15d\n",
			i*config.UpdateSize,
			serverOfflineTime.Microseconds(),
			(clientUpdateTime - serverOfflineTime).Microseconds(),
			offlineBytes,
			0,
			serverOnlineTime.Microseconds(),
			(clientReadTime - serverOnlineTime).Microseconds(),
			onlineBytes)

		if config.Progress {
			fmt.Fprintf(os.Stderr, "%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, config.NumUpdates)
		}
	}
}
