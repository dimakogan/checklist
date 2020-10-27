// +build BenchmarkIncremental

package boosted

import (
	"fmt"
	"math"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	for _, config := range testConfigs() {
		fmt.Printf("BenchmarkIncremental/" + config.String())
		result := testing.Benchmark(func(b *testing.B) {
			driver, err := ServerDriver()
			assert.NilError(b, err)
			rand := RandSource()

			var none int
			assert.NilError(b, driver.Configure(config, &none))
			client := NewPirClientUpdatable(RandSource(), [2]PirServer{driver, driver})

			err = client.Init()
			assert.NilError(b, err)

			driver.ResetMetrics(0, &none)
			var clientUpdateTime, clientReadTime time.Duration

			for i := 0; i < b.N; i++ {
				changeBatchSize := int(math.Round(math.Sqrt(float64(config.NumRows)))) + 1
				assert.NilError(b, driver.AddRows(changeBatchSize, &none))
				assert.NilError(b, driver.DeleteRows(changeBatchSize, &none))

				start := time.Now()
				client.Update()
				clientUpdateTime += time.Since(start)

				var rowIV RowIndexVal
				assert.NilError(b, driver.GetRow(rand.Intn(config.NumRows), &rowIV))

				start = time.Now()
				row, err := client.Read(int(rowIV.Key))
				clientReadTime += time.Since(start)
				assert.NilError(b, err)
				assert.DeepEqual(b, row, rowIV.Value)

				if *progress {
					fmt.Printf("%4d/%-5d\b\b\b\b\b\b\b\b\b\b", i, b.N)
				}
			}
			var serverHintTime, serverAnswerTime time.Duration
			assert.NilError(b, driver.GetHintTimer(0, &serverHintTime))
			assert.NilError(b, driver.GetAnswerTimer(0, &serverAnswerTime))

			b.ReportMetric(float64(serverHintTime.Nanoseconds())/float64(b.N), "hint-ns/op")
			b.ReportMetric(float64(serverAnswerTime.Nanoseconds())/float64(b.N), "answer-ns/op")
			b.ReportMetric(float64(clientUpdateTime.Nanoseconds())/float64(b.N), "update-ns/op")
			b.ReportMetric(float64(clientReadTime.Nanoseconds())/float64(b.N), "read-ns/op")
		})
		fmt.Printf("%s\n", result.String())
	}
}
