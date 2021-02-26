package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	. "github.com/dimakogan/boosted-pir"

	"gotest.tools/assert"
)

const (
	ColumnTimestamp = 0
	ColumnAdds      = 1
	ColumnDeletes   = 2
	ColumnQueries   = 3
)

func loadTraceFile(filename string) [][]int {
	var trace [][]int
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open trace file %s: %s", file, err)
	}

	r := csv.NewReader(file)
	r.Comment = '#'
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for row := range records {
		ts, err := strconv.Atoi(records[row][ColumnTimestamp])
		if err != nil {
			log.Fatalf("Bad row #%d timestamp: %s", row, records[row][ColumnTimestamp])
		}

		adds, err := strconv.Atoi(records[row][ColumnAdds])
		if err != nil {
			log.Fatalf("Bad row #%d adds: %s", row, records[row][ColumnAdds])
		}
		deletes, err := strconv.Atoi(records[row][ColumnDeletes])
		if err != nil {
			log.Fatalf("Bad row #%d deletes: %s", row, records[row][ColumnDeletes])
		}
		queries, err := strconv.Atoi(records[row][ColumnQueries])
		if err != nil {
			log.Fatalf("Bad row #%d deletes: %s", row, records[row][ColumnQueries])
		}
		if adds+deletes+queries > 0 {
			trace = append(trace, []int{ts, adds, deletes, queries})
		}
	}

	return trace
}

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

	rand := RandSource()

	var trace [][]int
	trace = loadTraceFile(config.TraceFile)
	config.NumRows = 0

	var none int
	if err := driver.Configure(config.TestConfig, &none); err != nil {
		log.Fatalf("Failed to configure driver: %s\n", err)
	}
	driver.ResetMetrics(0, &none)

	client := NewPirClientUpdatable(RandSource(), config.PirType, [2]PirUpdatableServer{driver, driver})

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
			var numRows int
			assert.NilError(ep, driver.NumRows(none, &numRows))
			assert.NilError(ep, driver.GetRow(rand.Intn(numRows), &rowIV))

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
			storageBytes = client.StorageNumBytes()
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

	// fmt.Fprintf(os.Stderr, "# NumLayerActivations: %v\n", NumLayerActivations)
	// fmt.Fprintf(os.Stderr, "# NumLayerHintBytes: %v\n", NumLayerHintBytes)
}
