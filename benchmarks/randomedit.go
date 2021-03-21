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

var grammar = flag.String("grammar", "", "grammar to use")
var input = flag.String("input", "", "input file to parse")
var edits = flag.Int("edits", 100, "number of edits")
var threshold = flag.Int("threshold", 4096, "memoization threshold")

func main() {
	flag.Parse()

	peg, err := ioutil.ReadFile(*grammar)
	if err != nil {
		log.Fatal(err)
	}

	p := re.MustCompilePatt(string(peg))
	prog := pattern.MustCompile(p)
	code := vm.Encode(prog)

	data, err := ioutil.ReadFile(*input)
	if err != nil {
		log.Fatal(err)
	}
	in := bytes.NewReader(data)

	tbl := memo.NewTreeTable(*threshold)
	tstart := time.Now()
	match, n, ast, _ := code.Exec(in, tbl)
	fmt.Println("elapsed:", time.Since(tstart))
	fmt.Printf("match: %t, n: %d, len(ast): %d\n", match, n, len(ast))

	for i := 0; i < *edits; i++ {
		pos := rand.Intn(len(data))
		e := memo.Edit{
			Start: pos,
			End:   pos + 1,
			Len:   1,
		}
		tstart := time.Now()
		tbl.ApplyEdit(e)
		match, n, ast, _ := code.Exec(in, tbl)
		fmt.Println("elapsed:", time.Since(tstart))
		fmt.Printf("match: %t, n: %d, len(ast): %d\n", match, n, len(ast))
	}
}
