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
		for i := 0; i < 1; i++ {
			code.Exec(in, tbl)
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
			fmt.Println(len(data), time.Since(tstart).Microseconds())
		}
		// f, err := os.Create("out.svg")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer f.Close()
		// viz.DrawMemo(tbl, len(data), f, 1000, 2000)
	}
	// fmt.Println("elapsed:", time.Since(tstart))
	// fmt.Printf("match: %t, n: %d, len(ast): %d\n", match, n, len(ast))
	// fmt.Printf("table entries: %d\n", tbl.Size())
}
