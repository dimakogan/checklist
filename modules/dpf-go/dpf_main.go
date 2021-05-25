package main

import (
	"flag"
	"fmt"
	"github.com/dkales/dpf-go/dpf"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpuprofile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	a, _ := dpf.Gen(123, 27)
	evalStart := time.Now()
	for i := 0; i < 100; i++ {
		dpf.EvalFull(a, 27)
	}
	fmt.Println("EvalFull time", time.Since(evalStart))
}
