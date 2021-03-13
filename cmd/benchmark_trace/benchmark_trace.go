package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	. "github.com/dimakogan/boosted-pir/driver"
	"github.com/dimakogan/boosted-pir/pir"
	"github.com/dimakogan/boosted-pir/updatable"

	"gotest.tools/assert"
)

func main() {
	var ep ErrorPrinter
	var numUpdates int

	config := new(Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()

	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	fmt.Printf("%15s%15s%15s%15s%15s%15s%15s%15s\n",
		"Timestamp",
		"NumAdds",
		"NumDeleted",
		"NumQueries",
		"ServerTime[us]", "ClientTime[us]", "CommBytes", "ClientStorage")

	var storageBytes int
	driver, err := config.ServerDriver()
	if err != nil {
		log.Fatalf("Failed to create driver: %s\n", err)
	}

	rand := pir.RandSource()

	var trace [][]int
	trace = LoadTraceFile(config.TraceFile)
	config.NumRows = 0

	var none int
	if err := driver.Configure(config.TestConfig, &none); err != nil {
		log.Fatalf("Failed to configure driver: %s\n", err)
	}
	driver.ResetMetrics(0, &none)

	client := updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{driver, driver})

	var clientTime, serverTime time.Duration
	var numBytes int
	numUpdates = len(trace) - 1

	for i := 0; i < numUpdates; i++ {
		ts := trace[i][ColumnTimestamp]
		numAdds := trace[i][ColumnAdds]
		numDeletes := trace[i][ColumnDeletes]
		numQueries := trace[i][ColumnQueries]

		if numAdds+numDeletes > 0 {
			assert.NilError(ep, driver.AddRows(numAdds, &none))
			assert.NilError(ep, driver.DeleteRows(numDeletes, &none))

			driver.ResetMetrics(0, &none)

			start := time.Now()
			client.Update()
			clientTime = time.Since(start)

			assert.NilError(ep, driver.GetOfflineTimer(0, &serverTime))
			assert.NilError(ep, driver.GetOfflineBytes(0, &numBytes))

		}

		if numQueries > 0 {
			var rowIV RowIndexVal
			var numKeys int
			assert.NilError(ep, driver.NumKeys(none, &numKeys))
			assert.NilError(ep, driver.GetRow(rand.Intn(numKeys), &rowIV))

			driver.ResetMetrics(0, &none)

			start := time.Now()
			row, err := client.Read(rowIV.Key)
			clientTime = time.Since(start)
			assert.NilError(ep, err)
			assert.DeepEqual(ep, row, rowIV.Value)

			assert.NilError(ep, driver.GetOnlineTimer(0, &serverTime))
			assert.NilError(ep, driver.GetOnlineBytes(0, &numBytes))

		}

		if i%200 == 0 {
			storageBytes = client.StorageNumBytes(SerializedSizeOf)
		}

		fmt.Printf("%15d%15d%15d%15d%15d%15d%15d%15d\n",
			ts,
			numAdds,
			numDeletes,
			numQueries,
			serverTime.Microseconds(),
			(clientTime - serverTime).Microseconds(),
			numBytes,
			storageBytes)

		if config.Progress {
			fmt.Fprintf(os.Stderr, "%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, numUpdates)
		}
	}
}
