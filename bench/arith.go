package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func arith() {
	p := Grammar("Expr", map[string]Pattern{
		"Expr":   Memo(Concat(NonTerm("Factor"), Star(Concat(Set(charset.New([]byte{'+', '-'})), NonTerm("Factor"))))),
		"Factor": Memo(Concat(NonTerm("Term"), Star(Concat(Set(charset.New([]byte{'*', '/'})), NonTerm("Term"))))),
		"Term":   Memo(Or(NonTerm("Number"), Concat(Concat(Literal("("), NonTerm("Expr")), Literal(")")))),
		"Number": Memo(Plus(Set(charset.Range('0', '9')))),
	})

	p.Optimize()
	fmt.Println(p)
	code := vm.Encode(p)
	fmt.Println(code)

	data, err := ioutil.ReadFile("../testdata/arith.txt")
	if err != nil {
		log.Fatal(err)
	}
	in := input.ByteReader(input.ByteReader(data))
	machine := vm.NewVM(in, code)

	fmt.Printf("Input is %d bytes\n", len(data))
	PrintMemUsage()

	tbl := memo.NewMapTable()
	// tbl := memo.NewLRUTable(100)
	// tbl := memo.NoneTable{}
	start := time.Now()
	fmt.Println(machine.Exec(tbl))
	elapsed := time.Since(start)

	fmt.Println("Parse completed in", elapsed)
	fmt.Printf("Memo table has %d entries\n", tbl.Size())
	start = time.Now()
	tbl.ApplyEdit(memo.Edit{
		Start: 0,
		End:   1,
		Len:   1,
	})
	elapsed = time.Since(start)
	fmt.Println("ApplyEdit in", elapsed)
	fmt.Printf("Memo table has %d entries\n", tbl.Size())

	machine.Reset()
	machine.SeekTo(0)
	start = time.Now()
	fmt.Println(machine.Exec(tbl))
	elapsed = time.Since(start)

	fmt.Println("Parse completed in", elapsed)
	PrintMemUsage()
}
