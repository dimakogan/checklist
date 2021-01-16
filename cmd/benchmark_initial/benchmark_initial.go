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
	"testing"
	"time"

	. "github.com/dimakogan/boosted-pir"
	"gotest.tools/assert"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	InitTestFlags()

	var ep ErrorPrinter

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
	fmt.Printf("%10s%22s%22s%15s%15s%22s%22s%15s\n",
		"numRows", "OfflineServerTime[us]", "OfflineClientTime[us]", "OfflineBytes", "ClientBytes",
		"OnlineServerTime[us]", "OnlineClientTime[us]", "OnlineBytes")

	for _, config := range TestConfigs() {
		driver, err := ServerDriver()
		if err != nil {
			log.Fatalf("Failed to create driver: %s\n", err)
		}

		rand := RandSource()

		var clientStatic PirClient
		var clientUpdatable PirUpdatableClient
		var none int
		if err := driver.Configure(config, &none); err != nil {
			log.Fatalf("Failed to configure driver: %s\n", err)
		}

		result := testing.Benchmark(func(b *testing.B) {
			assert.NilError(ep, driver.ResetMetrics(0, &none))
			var clientInitTime time.Duration
			var clientBytes int
			for i := 0; i < b.N; i++ {
				start := time.Now()
				if config.Updatable {
					clientUpdatable = NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})
					err = clientUpdatable.Init()
				} else {
					clientStatic = NewPIRClient(NewPirClientByType(config.PirType, rand), rand,
						[2]PirServer{driver, driver})
					err = clientStatic.Init()

				}
				assert.NilError(ep, err)
				clientInitTime += time.Since(start)
				if config.Updatable {
					clientBytes += clientUpdatable.StorageNumBytes()
				}
			}

			var serverHintTime time.Duration
			assert.NilError(ep, driver.GetHintTimer(0, &serverHintTime))
			b.ReportMetric(float64(serverHintTime.Microseconds())/float64(b.N), "hint-us/op")
			b.ReportMetric(float64((clientInitTime-serverHintTime).Microseconds())/float64(b.N), "init-us/op")

			var hintBytes int
			assert.NilError(ep, driver.GetHintBytes(0, &hintBytes))
			b.ReportMetric(float64(hintBytes)/float64(b.N), "hint-bytes/op")
			b.ReportMetric(float64(clientBytes)/float64(b.N), "client-bytes/op")
		})
		fmt.Printf("%10d%22d%22d%15d%15d",
			config.NumRows,
			int(result.Extra["hint-us/op"]),
			int(result.Extra["init-us/op"]),
			int(result.Extra["hint-bytes/op"]),
			int(result.Extra["client-bytes/op"]))

		result = testing.Benchmark(func(b *testing.B) {
			assert.NilError(ep, driver.ResetMetrics(0, &none))
			var clientReadTime time.Duration
			for i := 0; i < b.N; i++ {
				var rowIV RowIndexVal
				var row Row

				assert.NilError(ep, driver.GetRow(rand.Intn(config.NumRows), &rowIV))

				start := time.Now()
				if clientStatic != nil {
					row, err = clientStatic.Read(rowIV.Index)
				} else {
					row, err = clientUpdatable.Read(rowIV.Key)
				}
				clientReadTime += time.Since(start)
				assert.NilError(ep, err)
				assert.DeepEqual(ep, row, rowIV.Value)

				if i == b.N-2 {
					runtime.GC()
					if memProf, err := os.Create("mem.prof"); err != nil {
						panic(err)
					} else {
						pprof.WriteHeapProfile(memProf)
						memProf.Close()
					}
				}
			}
			var serverAnswerTime time.Duration
			assert.NilError(ep, driver.GetAnswerTimer(0, &serverAnswerTime))
			b.ReportMetric(float64(serverAnswerTime.Microseconds())/float64(b.N), "answer-us/op")
			b.ReportMetric(float64((clientReadTime-serverAnswerTime).Microseconds())/float64(b.N), "read-us/op")

			var answerBytes int
			assert.NilError(ep, driver.GetAnswerBytes(0, &answerBytes))
			b.ReportMetric(float64(answerBytes)/float64(b.N), "answer-bytes/op")

		})
		fmt.Printf("%22d%22d%15d\n",
			int(result.Extra["answer-us/op"]),
			int(result.Extra["read-us/op"]),
			int(result.Extra["answer-bytes/op"]))
	}
}
