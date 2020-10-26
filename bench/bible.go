package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func searchLast(patt Pattern) Pattern {
	p := Search(patt)
	// p := Grammar("S", map[string]Pattern{
	// 	"S": Or(patt, Concat(Any(1), NonTerm("S"))),
	// })
	return Star(p)
}

func bible() {
	var err error
	var bible input.ByteReader
	bible, err = ioutil.ReadFile("../testdata/bible.txt")
	if err != nil {
		fmt.Println("Warning:", err)
	}

	p := Concat(Plus(Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')))), Literal(" Abram"))
	// p := Literal("eartt")
	p = searchLast(p)
	p.Optimize()
	fmt.Println(p)
	code := vm.Encode(p)
	machine := vm.NewVM(bible, code)

	start := time.Now()
	match, _, _ := machine.Exec(memo.NoneTable{})
	fmt.Println(time.Since(start))
	fmt.Println(match)
}
