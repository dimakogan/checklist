package boosted

import (
	"log"
	"os"
	"runtime/pprof"
)

type Profiler struct {
	f *os.File
}

func NewCPUProfiler(filename string) *Profiler {
	prof := new(Profiler)
	if filename != "" {
		var err error
		prof.f, err = os.Create(filename)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(prof.f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
	return prof
}

func (p *Profiler) Close() {
	if p.f == nil {
		return
	}
	pprof.StopCPUProfile()
	p.f.Close()
}
