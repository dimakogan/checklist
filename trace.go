package boosted

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

const (
	ColumnTimestamp = 0
	ColumnAdds      = 1
	ColumnDeletes   = 2
	ColumnQueries   = 3
)

func LoadTraceFile(filename string) [][]int {
	var trace [][]int
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open trace file %s: %s", filename, err)
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
