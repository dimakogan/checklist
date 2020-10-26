// +build BenchmarkInitial

package boosted

import (
	"fmt"
	"log"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
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

		fmt.Printf("Init/" + config.String())
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
			b.ReportMetric(float64(serverHintTime.Nanoseconds())/float64(b.N), "hint-ns/op")
			b.ReportMetric(float64(clientInitTime.Nanoseconds())/float64(b.N), "init-ns/op")
		})
		fmt.Printf("%s\n", result.String())

		fmt.Printf("Read/" + config.String())
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
			b.ReportMetric(float64(serverAnswerTime.Nanoseconds())/float64(b.N), "answer-ns/op")
			b.ReportMetric(float64(clientReadTime.Nanoseconds())/float64(b.N), "read-ns/op")
		})
		fmt.Printf("%s\n", result.String())
	}
}
