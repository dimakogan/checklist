// +build BenchmarkInitial

package boosted

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	fmt.Printf("# go test -tags=BenchmarkInitial %s\n", strings.Join(os.Args[1:], " "))
	fmt.Printf("numRows\tHintServerTime[us]\tHintClientTime[us]\tReadServerTime[us]\tReadClientTime[us]\n")

	for _, config := range testConfigs() {
		driver, err := ServerDriver()
		if err != nil {
			log.Fatalf("Failed to create driver: %s\n", err)
		}

		rand := RandSource()

		var client PirClient
		var clientInitTime, clientReadTime time.Duration
		var none int
		if err := driver.Configure(config, &none); err != nil {
			log.Fatalf("Failed to configure driver: %s\n", err)
		}

		result := testing.Benchmark(func(b *testing.B) {
			assert.NilError(b, driver.ResetTimers(0, nil))
			for i := 0; i < b.N; i++ {
				if config.Updatable {
					client = NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})
				} else {
					client = NewPIRClient(NewPirClientByType(config.PirType, rand), rand,
						[2]PirServer{driver, driver})
				}
				start := time.Now()
				err = client.Init()
				assert.NilError(b, err)
				clientInitTime += time.Since(start)
			}

			var serverHintTime time.Duration
			assert.NilError(b, driver.GetHintTimer(0, &serverHintTime))
			b.ReportMetric(float64(serverHintTime.Microseconds())/float64(b.N), "hint-us/op")
			b.ReportMetric(float64((clientInitTime-serverHintTime).Microseconds())/float64(b.N), "init-us/op")
		})
		fmt.Printf("%d\t%d\t%d\t",
			config.NumRows,
			int(result.Extra["hint-us/op"]),
			int(result.Extra["init-us/op"]))

		result = testing.Benchmark(func(b *testing.B) {
			assert.NilError(b, driver.ResetTimers(0, nil))
			for i := 0; i < b.N; i++ {
				var rowIV RowIndexVal
				assert.NilError(b, driver.GetRow(rand.Intn(config.NumRows), &rowIV))

				start := time.Now()
				row, err := client.Read(int(rowIV.Key))
				clientReadTime += time.Since(start)
				assert.NilError(b, err)
				assert.DeepEqual(b, row, rowIV.Value)
			}
			var serverAnswerTime time.Duration
			assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))
			b.ReportMetric(float64(serverAnswerTime.Microseconds())/float64(b.N), "answer-us/op")
			b.ReportMetric(float64((clientReadTime-serverAnswerTime).Microseconds())/float64(b.N), "read-us/op")
		})
		fmt.Printf("%d\t%d\n",
			int(result.Extra["answer-us/op"]),
			int(result.Extra["read-us/op"]))
	}
}
