package driver

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

type Profiler struct {
	f        *os.File
	filename string
}

func NewProfiler(filename string) *Profiler {
	prof := new(Profiler)
	prof.filename = filename
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

	runtime.GC()
	if memProf, err := os.Create(p.filename + "-mem.prof"); err != nil {
		panic(err)
	} else {
		pprof.WriteHeapProfile(memProf)
		memProf.Close()
	}
}
