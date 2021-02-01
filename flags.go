package boosted

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var numRows string
var rowLen string
var pirType string
var updatable bool
var serverAddr string
var useTLS bool
var updateSize string

func InitTestFlags() {
	flag.StringVar(&numRows, "numRows", "10000", "Num DB Rows (comma-separated list)")
	flag.StringVar(&rowLen, "rowLen", "32", "Row length in bytes (comma-separated list)")
	flag.StringVar(&pirType, "pirType", Punc.String(),
		fmt.Sprintf("Updatable PIR type: [%s] (comma-separated list)", strings.Join(PirTypeStrings(), "|")))
	flag.BoolVar(&updatable, "updatable", true, "Test Updatable PIR")
	flag.StringVar(&serverAddr, "serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")
	flag.BoolVar(&useTLS, "tls", true, "Should use TLS")
	flag.StringVar(&updateSize, "updateSize", "1000", "number of rows in each update batch (default: 1000)")

	flag.Parse()

	fmt.Fprintf(os.Stderr, "# TestConfig: %v\n", TestConfigs())
}

func TestConfigs() []TestConfig {
	var configs []TestConfig
	numRowsStr := strings.Split(numRows, ",")
	dbRowLenStr := strings.Split(rowLen, ",")
	pirTypeStrs := strings.Split(pirType, ",")
	updateSizeStrs := strings.Split(updateSize, ",")

	for _, nStr := range numRowsStr {
		n, err := strconv.Atoi(nStr)
		if err != nil {
			log.Fatalf("Bad numRows: %s\n", nStr)
		}
		for _, rowLenStr := range dbRowLenStr {
			recSize, err := strconv.Atoi(rowLenStr)
			if err != nil {
				log.Fatalf("Bad rowLen: %s\n", rowLenStr)
			}

			for _, pirTypeStr := range pirTypeStrs {
				pirType, err := PirTypeString(pirTypeStr)
				if err != nil {
					log.Fatalf("Bad PirType: %s\n", pirTypeStr)
				}
				config := TestConfig{NumRows: n, RowLen: recSize, PirType: pirType, Updatable: updatable}
				if pirType == Perm {
					config.NumRows = 1 << int(math.Ceil(math.Log2(float64(config.NumRows))))
				}
				for _, updateSizeStr := range updateSizeStrs {
					updateSize, err := strconv.Atoi(updateSizeStr)
					if err != nil {
						log.Fatalf("Bad updateSize: %s\n", updateSizeStr)
					}
					config.UpdateSize = updateSize
					configs = append(configs, config)
				}
			}
		}

	}

	return configs
}

func ServerDriver() (PirServerDriver, error) {
	if serverAddr != "" {
		return NewPirRpcProxy(serverAddr, useTLS)
	} else {
		return NewPirServerDriver()
	}
}
