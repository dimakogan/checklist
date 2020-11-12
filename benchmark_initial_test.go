// +build BenchmarkInitial

package boosted

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
)

var cpuprof = flag.String("cpuprof", "", "write cpu profile to `file`")

func TestMain(m *testing.M) {
	flag.Parse()
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("# go test -tags=BenchmarkInitial %s\n", strings.Join(os.Args[1:], " "))
	fmt.Printf("%10s%22s%22s%15s%15s%22s%22s%15s\n",
		"numRows", "OfflineServerTime[us]", "OfflineClientTime[us]", "OfflineBytes", "ClientBytes",
		"OnlineServerTime[us]", "OnlineClientTime[us]", "OnlineBytes")

	for _, config := range testConfigs() {
		driver, err := ServerDriver()
		if err != nil {
			log.Fatalf("Failed to create driver: %s\n", err)
		}

		rand := RandSource()

		var client PirClient
		var none int
		if err := driver.Configure(config, &none); err != nil {
			log.Fatalf("Failed to configure driver: %s\n", err)
		}

		result := testing.Benchmark(func(b *testing.B) {
			assert.NilError(ep, driver.ResetMetrics(0, nil))
			var clientInitTime time.Duration
			var clientBytes uint64
			for i := 0; i < b.N; i++ {
				var m1, m2 runtime.MemStats
				runtime.GC()
				runtime.ReadMemStats(&m1)
				if config.Updatable {
					client = NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})
				} else {
					client = NewPIRClient(NewPirClientByType(config.PirType, rand), rand,
						[2]PirServer{driver, driver})
				}
				start := time.Now()
				err = client.Init()
				assert.NilError(ep, err)
				clientInitTime += time.Since(start)
				runtime.GC()
				runtime.ReadMemStats(&m2)
				clientBytes += m2.Alloc - m1.Alloc
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
			assert.NilError(ep, driver.ResetMetrics(0, nil))
			var clientReadTime time.Duration
			for i := 0; i < b.N; i++ {
				var rowIV RowIndexVal
				assert.NilError(ep, driver.GetRow(-1 /* random row */, &rowIV))

				start := time.Now()
				row, err := client.Read(int(rowIV.Key))
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
