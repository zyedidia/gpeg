package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/dustin/go-humanize"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var bench = flag.String("b", "", "benchmark to run")

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v", humanize.Bytes(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v", humanize.Bytes(m.TotalAlloc))
	fmt.Printf("\tSys = %v", humanize.Bytes(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *bench == "" {
		fmt.Println("No benchmark given")
		os.Exit(1)
	}

	switch *bench {
	case "arith":
		arith()
	case "bible":
		bible()
	case "json":
		json()
	default:
		fmt.Println("Unknown benchmark", *bench)
		os.Exit(1)
	}
}
