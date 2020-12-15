package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang-collections/go-datastructures/augmentedtree"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/pegexp"
	"github.com/zyedidia/gpeg/vm"
)

func fatal(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}

func searchLast(p pattern.Pattern) pattern.Pattern {
	return pattern.Star(pattern.Search(pattern.Cap(p)))
}

var peg = flag.String("peg", "", "file to load peg from")
var glob = flag.String("glob", "", "only search files that match this glob")
var showsearch = flag.Bool("showsearch", false, "display the program used for searching")
var showpatt = flag.Bool("showpatt", false, "display the program using for matching")

func main() {
	flag.Parse()
	args := flag.Args()

	var target pattern.Pattern
	if *peg != "" {
		data, err := ioutil.ReadFile(*peg)
		if err != nil {
			fatal(err)
		}
		target, err = pegexp.CompilePatt(string(data))
		if err != nil {
			fatal(err)
		}
	} else {
		if len(args) < 1 {
			fatal("too few arguments")
		}
		var err error
		target, err = pegexp.CompilePatt(args[0])
		if err != nil {
			fatal(err)
		}
	}
	search, err := pattern.Compile(searchLast(target))
	if err != nil {
		fatal(err)
	}

	if *showpatt {
		fmt.Println(pattern.MustCompile(target))
	}

	if *showsearch {
		fmt.Println(search)
	}

	in, err := os.Open("test_longer.xml")
	if err != nil {
		fatal(err)
	}
	fi, err := in.Stat()
	if err != nil {
		fatal(err)
	}
	sz := fi.Size()

	data, lines, err := read(in, sz)
	if err != nil {
		fatal(err)
	}

	code := vm.Encode(search)
	machine := vm.NewVM(input.ByteReader(data), code)
	_, _, ast := machine.Exec(memo.NoneTable{})
	fmt.Println("done matching")
	caps := machine.CapturesIndex(ast)

	tree := augmentedtree.New(1)
	tree.Add(lines...)
	for _, c := range caps {
		start := tree.Query(Line{
			Start: int(c[0]),
			End:   int(c[0]),
		})
		end := tree.Query(Line{
			Start: int(c[1]),
			End:   int(c[1]),
		})
		text := string(data[start[0].LowAtDimension(0):end[0].HighAtDimension(0)])
		fmt.Printf("%d: %s\n", start[0].ID(), text)
	}
}
