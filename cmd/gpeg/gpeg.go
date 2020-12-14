package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/pegexp"
	"github.com/zyedidia/gpeg/viz"
	"github.com/zyedidia/gpeg/vm"
)

var peg = flag.String("peg", "", "PEG input grammar")
var out = flag.String("o", "", "Output file")

func fatal(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}

func main() {
	var src string
	var srcname string

	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		src = args[0]
		srcname = filepath.Base(src)
		srcname = strings.TrimSuffix(srcname, filepath.Ext(srcname))
	}

	if *peg == "" {
		fatal("error: no PEG")
	}

	pegname := filepath.Base(*peg)
	pegname = strings.TrimSuffix(pegname, filepath.Ext(pegname))

	data, err := ioutil.ReadFile(*peg)
	if err != nil {
		fatal(err)
	}
	p, err := pegexp.CompilePatt(string(data))
	if err != nil {
		fatal(err)
	}

	if src != "" {
		peg, err := pattern.Compile(p)
		if err != nil {
			fatal(err)
		}
		code := vm.Encode(peg)
		srcdata, err := ioutil.ReadFile(src)
		if err != nil {
			fatal(err)
		}
		in := input.ByteReader(srcdata)
		machine := vm.NewVM(in, code)
		match, _, ast := machine.Exec(memo.NoneTable{})
		if !match {
			fatal("Parse failed")
		}
		outf := srcname + ".pdf"
		if *out != "" {
			outf = *out
		}
		viz.WriteDotViz(outf, viz.GraphAST(ast, map[string]int16{}))
	} else {
		g, ok := p.(*pattern.GrammarNode)
		if !ok {
			fatal("error: top-level node is not a grammar")
		}
		g.Inline()

		outf := pegname + ".pdf"
		if *out != "" {
			outf = *out
		}

		viz.WriteDotViz(outf, viz.Graph(g))
	}
}
