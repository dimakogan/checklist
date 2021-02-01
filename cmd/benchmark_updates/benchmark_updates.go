package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"strings"
	"time"

	. "github.com/dimakogan/boosted-pir"

	"gotest.tools/assert"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var progress = flag.Bool("progress", true, "Show benchmarks progress")

func main() {
	NumLayerActivations = make(map[int]int)
	NumLayerHintBytes = make(map[int]int)

	var ep ErrorPrinter
	var numUpdates int
	flag.IntVar(&numUpdates, "numUpdates", 0, "number of update batches (default: numRows/updateSize)")

	InitTestFlags()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	fmt.Printf("%15s%22s%22s%22s%15s%22s%22s%15s\n",
		"NumUpdates",
		"UpdateServerTime[us]", "UpdateClientTime[us]", "UpdateBytes", "ClientBytes",
		"OnlineServerTime[us]", "OnlineClientTime[us]", "OnlineBytes")

	for _, config := range TestConfigs() {
		driver, err := ServerDriver()
		if err != nil {
			log.Fatalf("Failed to create driver: %s\n", err)
		}

		rand := RandSource()

		var none int
		if err := driver.Configure(config, &none); err != nil {
			log.Fatalf("Failed to configure driver: %s\n", err)
		}
		driver.ResetMetrics(0, &none)

		client := NewPirClientUpdatable(RandSource(), config.PirType, [2]PirUpdatableServer{driver, driver})

		var clientUpdateTime, clientReadTime time.Duration

		if numUpdates == 0 {
			numUpdates = (config.NumRows-1)/config.UpdateSize + 1
		}
		for i := 0; i < numUpdates; i++ {
			assert.NilError(ep, driver.AddRows(config.UpdateSize/2, &none))
			assert.NilError(ep, driver.DeleteRows(config.UpdateSize/2, &none))

			driver.ResetMetrics(0, &none)

			start := time.Now()
			client.Update()
			clientUpdateTime = time.Since(start)

			var rowIV RowIndexVal
			var numRows int
			assert.NilError(ep, driver.NumRows(none, &numRows))
			assert.NilError(ep, driver.GetRow(rand.Intn(numRows), &rowIV))

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

			if *progress {
				fmt.Fprintf(os.Stderr, "%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, numUpdates)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "# NumLayerActivations: %v\n", NumLayerActivations)
	fmt.Fprintf(os.Stderr, "# NumLayerHintBytes: %v\n", NumLayerHintBytes)
}
