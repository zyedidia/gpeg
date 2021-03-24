package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

var grammar = flag.String("grammar", "../grammars/java.peg", "grammar to use")

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

		tbl := memo.NoneTable{}
		tstart := time.Now()
		code.Exec(in, tbl)
		fmt.Println(len(data), time.Since(tstart).Microseconds())
	}
	// fmt.Println("elapsed:", time.Since(tstart))
	// fmt.Printf("match: %t, n: %d, len(ast): %d\n", match, n, len(ast))
	// fmt.Printf("table entries: %d\n", tbl.Size())
}
