package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"checklist/driver"
	"checklist/pir"
	"checklist/updatable"

	"gotest.tools/assert"
)

func main() {
	config := new(driver.Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()

	var ep driver.ErrorPrinter

	prof := driver.NewProfiler(config.CpuProfile)
	defer prof.Close()

	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	fmt.Printf("%10s%22s%22s%15s%15s%22s%22s%15s\n",
		"numRows", "OfflineServerTime[us]", "OfflineClientTime[us]", "OfflineBytes", "ClientBytes",
		"OnlineServerTime[us]", "OnlineClientTime[us]", "OnlineBytes")

	dr, err := config.ServerDriver()
	if err != nil {
		log.Fatalf("Failed to create driver: %s\n", err)
	}

	rand := pir.RandSource()

	var clientStatic pir.PIRReader
	var clientUpdatable *updatable.Client
	var none int
	if err := dr.Configure(config.TestConfig, &none); err != nil {
		log.Fatalf("Failed to configure driver: %s\n", err)
	}

	result := testing.Benchmark(func(b *testing.B) {
		assert.NilError(ep, dr.ResetMetrics(0, &none))
		var clientInitTime time.Duration
		var clientBytes int
		for i := 0; i < b.N; i++ {
			start := time.Now()
			if config.Updatable {
				clientUpdatable = updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{dr, dr})
				err = clientUpdatable.Init()
			} else {
				clientStatic = pir.NewPIRReader(rand, dr, dr)
				err = clientStatic.Init(config.PirType)

			}
			assert.NilError(ep, err)
			clientInitTime += time.Since(start)
			if config.Updatable {
				clientBytes += clientUpdatable.StorageNumBytes(driver.SerializedSizeOf)
			}
		}

		var serverOfflineTime time.Duration
		assert.NilError(ep, dr.GetOfflineTimer(0, &serverOfflineTime))
		b.ReportMetric(float64(serverOfflineTime.Microseconds())/float64(b.N), "hint-us/op")
		b.ReportMetric(float64((clientInitTime-serverOfflineTime).Microseconds())/float64(b.N), "init-us/op")

		var offlineBytes int
		assert.NilError(ep, dr.GetOfflineBytes(0, &offlineBytes))
		b.ReportMetric(float64(offlineBytes)/float64(b.N), "hint-bytes/op")
		b.ReportMetric(float64(clientBytes)/float64(b.N), "client-bytes/op")
	})
	fmt.Printf("%10d%22d%22d%15d%15d",
		config.NumRows,
		int(result.Extra["hint-us/op"]),
		int(result.Extra["init-us/op"]),
		int(result.Extra["hint-bytes/op"]),
		int(result.Extra["client-bytes/op"]))

	result = testing.Benchmark(func(b *testing.B) {
		assert.NilError(ep, dr.ResetMetrics(0, &none))
		var clientReadTime time.Duration
		for i := 0; i < b.N; i++ {
			var rowIV driver.RowIndexVal
			var row pir.Row

			assert.NilError(ep, dr.GetRow(rand.Intn(config.NumRows), &rowIV))

			start := time.Now()
			if clientStatic != nil {
				row, err = clientStatic.Read(rowIV.Index)
			} else {
				row, err = clientUpdatable.Read(rowIV.Key)
			}
			clientReadTime += time.Since(start)
			assert.NilError(ep, err)
			if row[0] != rowIV.Value[0] {
				fmt.Printf("BAD: %d\n", i)
			}

			if i == b.N-2 {
				runtime.GC()
				if memProf, err := os.Create("mem.prof"); err != nil {
					log.Printf("Failed to create memory profile: %s", err)
				} else {
					pprof.WriteHeapProfile(memProf)
					memProf.Close()
				}
			}
		}
		var serverOnlineTime time.Duration
		assert.NilError(ep, dr.GetOnlineTimer(0, &serverOnlineTime))
		b.ReportMetric(float64(serverOnlineTime.Microseconds())/float64(b.N), "answer-us/op")
		b.ReportMetric(float64((clientReadTime-serverOnlineTime).Microseconds())/float64(b.N), "read-us/op")

		var onlineBytes int
		assert.NilError(ep, dr.GetOnlineBytes(0, &onlineBytes))
		b.ReportMetric(float64(onlineBytes)/float64(b.N), "answer-bytes/op")

	})
	fmt.Printf("%22d%22d%15d\n",
		int(result.Extra["answer-us/op"]),
		int(result.Extra["read-us/op"]),
		int(result.Extra["answer-bytes/op"]))

}
