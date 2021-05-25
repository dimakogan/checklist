package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	. "checklist/driver"
	"checklist/testing"
)

func main() {
	config := new(Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()
	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))

	if len(config.TraceFile) == 0 {
		log.Fatalf("Missing trace filename")
	}
	file, err := os.Open(config.TraceFile)
	if err != nil {
		log.Fatalf("Failed to open trace file %s: %s", config.TraceFile, err)
	}

	testing.BenchmarkTrace(config, file, os.Stdout, os.Stderr)
}
