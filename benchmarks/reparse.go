package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

var grammar = flag.String("grammar", "../grammars/json_memo.peg", "grammar to use")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {
	flag.Parse()

	inputs := flag.Args()

	peg, err := ioutil.ReadFile(*grammar)
	if err != nil {
		log.Fatal(err)
	}

	p := re.MustCompilePatt(string(peg))
	prog := pattern.MustCompile(p)
	code := vm.Encode(prog)

	for _, input := range inputs {
		data, err := ioutil.ReadFile(input)
		if err != nil {
			log.Fatal(err)
		}
		in := bytes.NewReader(data)

		tbl := memo.NewTreeTable(512)
		// tstart := time.Now()
		code.Exec(in, tbl)
		// fmt.Println(time.Since(tstart).Microseconds())

		// var total int64
		for i := 0; i < 1000; i++ {
			loc := rand.Intn(len(data))
			edit := memo.Edit{
				Start: loc,
				End:   loc + 1,
				Len:   1,
			}
			in.Reset(data)

			tstart := time.Now()
			tbl.ApplyEdit(edit)
			code.Exec(in, tbl)
			fmt.Println(time.Since(tstart).Microseconds())
			// var m runtime.MemStats
			// runtime.ReadMemStats(&m)
			// fmt.Printf("%d\n", bToKb(m.Alloc-uint64(len(data))))
			// total += time.Since(tstart).Microseconds()
		}
		// fmt.Println(len(data), int(float64(total)/10))
		// f, err := os.Create("out.svg")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer f.Close()
		// viz.DrawMemo(tbl, len(data), f, 1000, 2000)
	}
	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.WriteHeapProfile(f)
	// 	f.Close()
	// 	return
	// }
	// fmt.Println("elapsed:", time.Since(tstart))
	// fmt.Printf("match: %t, n: %d, len(ast): %d\n", match, n, len(ast))
	// fmt.Printf("table entries: %d\n", tbl.Size())
}
func bToKb(b uint64) uint64 {
	return b / 1024
}
