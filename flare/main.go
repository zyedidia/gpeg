package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fatih/color"
	"github.com/zyedidia/gpeg/memo"
	p "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
	"github.com/zyedidia/linerope"
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
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

var letters = []byte("\n \tabcdefghijklmnopqrstuvwxyz")

func randbytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
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

	prog := p.MustCompile(java)
	code := vm.Encode(prog)

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
	r := linerope.New(data, &linerope.DefaultOptions)

	tbl := memo.NewTreeTable(4096)
	// tbl := memo.NoneTable{}
	rand.Seed(42)
	// r := bytes.NewReader(data)
	istart := time.Now()
	match, n, ast, _ := code.Exec(r, tbl)
	ielapsed := time.Since(istart)
	fmt.Println("initial", ielapsed.Microseconds())

	var total int64
	var applyedit int64
	const nedits = 1

	if !*oneparse {
		for i := 0; i < nedits; i++ {
			text := randbytes(4)
			// text := []byte("n")
			// loc := rand.Intn(len(data))
			loc := rand.Intn(r.Len())
			for j := 0; j < 1; j++ {
				edit := memo.Edit{
					Start: loc,
					End:   loc,
					// End: loc + 1,
					// Len:   1,
					// Len:   1,
					// Len: 1,
					Len: len(text),
				}
				r.Insert(loc, text)
				loc += len(text)

				// data = append(data[:loc], append(text, data[loc:]...)...)

				astart := time.Now()
				tbl.ApplyEdit(edit)
				aelapsed := time.Since(astart)
				// r.Reset(data)
				match, n, ast, _ = code.Exec(r, tbl)
				telapsed := time.Since(astart)

				// fmt.Printf("%d %d\n", telapsed.Nanoseconds(), aelapsed.Nanoseconds())
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("%d %d\n", bToMb(m.Alloc), telapsed.Microseconds())

				total += telapsed.Nanoseconds()
				applyedit += aelapsed.Nanoseconds()
			}
		}
	}

	// fmt.Printf("%.3fus %.3fus\n", float64(total)/nedits/1000.0, float64(applyedit)/nedits/1000.0)

	if *display {
		buf := make([]byte, r.Len())
		n, _ := r.ReadAt(buf, int64(0))
		for _, c := range ast {
			fmt.Print(colorize(c, theme, buf[:n]))
		}
	}

	fmt.Println(match, n, len(ast))
	fmt.Println(tbl.Size())

	// f, err := os.Create("out.svg")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()
	// viz.DrawMemo(tbl, len(data), f, 1000, 2000)

	PrintMemUsage()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
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
