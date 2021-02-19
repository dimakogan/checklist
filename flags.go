package boosted

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

type Configurator struct {
	TestConfig

	UseTLS     bool
	CpuProfile string

	// For client
	ServerAddr    string
	ServerAddr2   string
	UsePersistent bool

	// For server
	Port int

	// For benchmarks
	NumUpdates int
	Progress   bool
	TraceFile  string

	pirTypeStr string

	FlagSet *flag.FlagSet
}

func NewConfig() *Configurator {
	c := new(Configurator)
	c.FlagSet = flag.CommandLine
	c.FlagSet.IntVar(&c.NumRows, "numRows", 10000, "Num DB Rows")
	c.FlagSet.IntVar(&c.RowLen, "rowLen", 32, "Row length in bytes")
	c.FlagSet.StringVar(&c.pirTypeStr, "pirType", Punc.String(),
		fmt.Sprintf("Updatable PIR type: [%s]", strings.Join(PirTypeStrings(), "|")))
	c.FlagSet.BoolVar(&c.Updatable, "updatable", true, "Test Updatable PIR")
	c.FlagSet.IntVar(&c.UpdateSize, "updateSize", 500, "number of rows in each update batch (default: 500)")
	c.FlagSet.StringVar(&c.CpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	return c
}

func (c *Configurator) WithClientFlags() *Configurator {
	c.FlagSet.StringVar(&c.ServerAddr, "serverAddr", "", "<HOSTNAME>:<PORT> of server for RPC test")
	c.FlagSet.StringVar(&c.ServerAddr2, "serverAddr2", "", "<HOSTNAME>:<PORT> of server for RPC test")
	c.FlagSet.BoolVar(&c.UseTLS, "tls", false, "Should use TLS")
	c.FlagSet.BoolVar(&c.UsePersistent, "persistent", false, "Should use peristent connection to server")
	return c
}

func (c *Configurator) WithServerFlags() *Configurator {
	c.FlagSet.BoolVar(&c.UseTLS, "tls", false, "Should use TLS")
	c.FlagSet.IntVar(&c.Port, "p", 12345, "Listening port")
	return c
}

func (c *Configurator) WithBenchmarkFlags() *Configurator {
	c.FlagSet.BoolVar(&c.Progress, "progress", true, "Show benchmarks progress")
	c.FlagSet.IntVar(&c.NumUpdates, "numUpdates", 0, "number of update batches (default: numRows/updateSize)")
	c.FlagSet.StringVar(&c.TraceFile, "trace", "trace.txt", "input trace file")
	return c
}

func (c *Configurator) Parse() *Configurator {
	if c.FlagSet.Parsed() {
		return c
	}
	if err := c.FlagSet.Parse(os.Args[1:]); err != nil {
		log.Fatalf("%v", err)
	}
	var err error
	c.PirType, err = PirTypeString(c.pirTypeStr)
	if err != nil {
		log.Fatalf("Bad PirType: %s\n", c.pirTypeStr)
	}
	if c.PirType == Perm {
		c.NumRows = 1 << int(math.Ceil(math.Log2(float64(c.NumRows))))
	}

	return c
}

func (c *Configurator) ServerDriver() (PirServerDriver, error) {
	c.Parse()

	if c.ServerAddr != "" {
		return NewPirRpcProxy(c.ServerAddr, c.UseTLS, c.UsePersistent)
	} else {
		return NewPirServerDriver()
	}
}

func (c *Configurator) Server2Driver() (PirServerDriver, error) {
	c.Parse()

	if c.ServerAddr2 != "" {
		return NewPirRpcProxy(c.ServerAddr2, c.UseTLS, c.UsePersistent)
	} else if c.ServerAddr != "" {
		return NewPirRpcProxy(c.ServerAddr, c.UseTLS, c.UsePersistent)
	} else {
		return NewPirServerDriver()
	}
}
