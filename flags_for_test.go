package boosted

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/rpc"
	"strconv"
	"strings"
)

var numRows = flag.String("numRows", "10000", "Num DB Rows (comma-separated list)")
var rowLen = flag.String("rowLen", "32", "Row length in bytes (comma-separated list)")
var pirType = flag.String("pirType", Punc.String(),
	fmt.Sprintf("Updatable PIR type: [%s] (comma-separated list)", strings.Join(PirTypeStrings(), "|")))
var serverAddr = flag.String("serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")

func testConfigs() []TestConfig {
	var configs []TestConfig
	numRowsStr := strings.Split(*numRows, ",")
	dbRowLenStr := strings.Split(*rowLen, ",")
	pirTypeStrs := strings.Split(*pirType, ",")

	// Set maximum on total size to avoid really large DBs.
	maxDBSizeBytes := int64(1 * 1024 * 1024 * 1024)

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

			if int64(n)*int64(recSize) > maxDBSizeBytes {
				continue
			}
			for _, pirTypeStr := range pirTypeStrs {
				pirType, err := PirTypeString(pirTypeStr)
				if err != nil {
					log.Fatalf("Bad PirType: %s\n", pirTypeStr)
				}
				config := TestConfig{NumRows: n, RowLen: recSize, PirType: pirType}
				if pirType == Perm {
					config.NumRows = 1 << int(math.Ceil(math.Log2(float64(config.NumRows))))
				}
				configs = append(configs, config)
			}
		}

	}

	return configs
}

func ServerDriver() (PirServerDriver, error) {
	if *serverAddr != "" {
		// Create a TCP connection to localhost on port 1234
		remote, err := rpc.DialHTTP("tcp", *serverAddr)
		if err != nil {
			return nil, err
		}
		return NewPirRpcProxy(remote), nil
	} else {
		return NewPirServerDriver()
	}
}
