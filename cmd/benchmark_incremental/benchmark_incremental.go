package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
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
	var numBatchesFlag int
	flag.IntVar(&numBatchesFlag, "numBatches", 0, "number of update batches (default: ~sqrt(numRows))")

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
	fmt.Printf("%10s%22s%22s%22s%15s%22s%22s%15s\n",
		"numRows",
		"UpdateServerTime[us]", "UpdateClientTime[us]", "UpdateBytesPerChange", "ClientBytes",
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

		client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

		err = client.Init()
		assert.NilError(ep, err)

		driver.ResetMetrics(0, &none)
		var clientUpdateTime, clientReadTime time.Duration

		changeBatchSize := 100000
		var numBatches int
		if numBatchesFlag == 0 {
			numBatches = (config.NumRows-1)/changeBatchSize + 1
		} else {
			numBatches = numBatchesFlag
		}

		for i := 0; i < numBatches; i++ {
			assert.NilError(ep, driver.AddRows(changeBatchSize, &none))
			assert.NilError(ep, driver.DeleteRows(changeBatchSize, &none))

			start := time.Now()
			client.Update()
			clientUpdateTime += time.Since(start)

			var rowIV RowIndexVal
			var numRows int
			assert.NilError(ep, driver.NumRows(none, &numRows))
			assert.NilError(ep, driver.GetRow(rand.Intn(numRows), &rowIV))

			start = time.Now()
			row, err := client.Read(int(rowIV.Key))
			clientReadTime += time.Since(start)
			assert.NilError(ep, err)
			assert.DeepEqual(ep, row, rowIV.Value)

			if *progress {
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
		var serverHintTime, serverAnswerTime time.Duration
		var hintBytes, answerBytes int
		assert.NilError(ep, driver.GetHintTimer(0, &serverHintTime))
		assert.NilError(ep, driver.GetAnswerTimer(0, &serverAnswerTime))
		assert.NilError(ep, driver.GetAnswerBytes(0, &answerBytes))
		assert.NilError(ep, driver.GetHintBytes(0, &hintBytes))

		fmt.Printf("%10d%22d%22d%22d%22d%22d%15d\n",
			config.NumRows,
			serverHintTime.Microseconds()/int64(numBatches),
			(clientUpdateTime-serverHintTime).Microseconds()/int64(numBatches),
			hintBytes/(numBatches*changeBatchSize*2),
			serverAnswerTime.Microseconds()/int64(numBatches),
			(clientReadTime-serverAnswerTime).Microseconds()/int64(numBatches),
			answerBytes/numBatches)
	}
	fmt.Fprintf(os.Stderr, "# NumLayerActivations: %v\n", NumLayerActivations)
	fmt.Fprintf(os.Stderr, "# NumLayerHintBytes: %v\n", NumLayerHintBytes)
}