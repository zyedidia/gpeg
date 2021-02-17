package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fatih/color"
	"github.com/zyedidia/gpeg/memo"
	p "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/viz"
	"github.com/zyedidia/gpeg/vm"
)

func colorize(c *memo.Capture, theme map[TokenType]*color.Color, text []byte) string {
	clr, ok := theme[TokenType(c.Id)]
	if !ok {
		ret := ""
		for _, ch := range c.Children {
			ret += colorize(ch, theme, text)
		}
		return ret
	}
	if len(c.Children) == 0 {
		return clr.Sprint(string(text[c.Start():c.End()]))
	}
	ret := ""
	last := c.Start()
	cend := 0
	for _, ch := range c.Children {
		ret += clr.Sprint(string(text[last:ch.Start()])) + colorize(ch, theme, text)
		cend = ch.End()
	}
	return ret + clr.Sprint(string(text[cend:c.End()]))
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	display    = flag.Bool("display", false, "display highlighted output")
	oneparse   = flag.Bool("oneparse", false, "only do initial parse")
)

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

	prog := p.MustCompile(java)
	code := vm.Encode(prog)

	fmt.Println("Size of instructions:", code.Size())
	codebytes, err := code.ToBytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialization size:", len(codebytes))
	code, err = vm.FromBytes(codebytes)
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	tbl := memo.NewTreeTable()
	// tbl := memo.NoneTable{}
	r := bytes.NewReader(data)
	istart := time.Now()
	match, n, ast, _ := code.Exec(r, tbl)
	ielapsed := time.Since(istart)
	fmt.Println("initial", ielapsed)

	if !*oneparse {
		text := "h"
		loc := 5
		edit := memo.Edit{
			Start: loc,
			End:   loc + 1,
			Len:   len(text) + 1,
		}

		data = append(data[:loc], append([]byte(text), data[loc:]...)...)

		astart := time.Now()
		tbl.ApplyEdit(edit)
		aelapsed := time.Since(astart)
		r.Reset(data)
		rstart := time.Now()
		match, n, ast, _ = code.Exec(r, tbl)
		relapsed := time.Since(rstart)

		if *display {
			for _, c := range ast {
				fmt.Print(colorize(c, theme, data))
			}
		}

		fmt.Printf("initial: %v, apply-edit: %v, reparse: %v\n", ielapsed, aelapsed, relapsed)
	}

	fmt.Println(match, n, len(ast))
	fmt.Println(tbl.Size())

	f, err := os.Create("out.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	viz.DrawMemo(tbl, len(data), f, 1000, 250)

	PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
