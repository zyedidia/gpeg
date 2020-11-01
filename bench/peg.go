package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func peg() {
	ids := make(map[string]int16)
	p := CapGrammar("Grammar", map[string]Pattern{
		"Grammar":    Concat(NonTerm("Spacing"), Plus(NonTerm("Definition")), NonTerm("EndOfFile")),
		"Definition": Concat(NonTerm("Identifier"), NonTerm("LEFTARROW"), NonTerm("Expression")),
		"Expression": Concat(NonTerm("Sequence"), Star(Concat(NonTerm("SLASH"), NonTerm("Sequence")))),
		"Sequence":   Star(NonTerm("Prefix")),
		"Prefix":     Concat(Optional(Or(NonTerm("AND"), NonTerm("NOT"))), NonTerm("Suffix")),
		"Suffix":     Concat(NonTerm("Primary"), Optional(Or(NonTerm("QUESTION"), NonTerm("STAR"), NonTerm("PLUS")))),
		"Primary": Or(Concat(NonTerm("Identifier"), Not(NonTerm("LEFTARROW"))),
			Concat(NonTerm("OPEN"), NonTerm("Expression"), NonTerm("CLOSE")),
			NonTerm("Literal"), NonTerm("Class"), NonTerm("DOT")),
		"Identifier": Concat(NonTerm("IdentStart"), Star(NonTerm("IdentCont")), NonTerm("Spacing")),
		"IdentStart": Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')).Add(charset.New([]byte{'_'}))),
		"IdentCont":  Or(NonTerm("IdentStart"), Set(charset.Range('0', '9'))),
		"Literal": Or(Concat(Literal("'"), Star(Concat(Not(Literal("'")), NonTerm("Char"))), Literal("'"), NonTerm("Spacing")),
			Concat(Literal("\""), Star(Concat(Not(Literal("\"")), NonTerm("Char"))), Literal("\""), NonTerm("Spacing"))),
		"Class": Concat(Literal("["), Star(Concat(Not(Literal("]")), NonTerm("Range"))), Literal("]"), NonTerm("Spacing")),
		"Range": Or(Concat(NonTerm("Char"), Literal("-"), NonTerm("Char")), NonTerm("Char")),
		"Char": Or(Concat(Literal("\\"), Set(charset.New([]byte{'n', 'r', 't', '\'', '"', '[', ']', '\\'}))),
			Concat(Literal("\\"), Set(charset.Range('0', '2')), Set(charset.Range('0', '7')), Set(charset.Range('0', '7'))),
			Concat(Literal("\\"), Set(charset.Range('0', '7')), Optional(Set(charset.Range('0', '7')))),
			Concat(Not(Literal("\\")), Any(1))),

		"LEFTARROW": Concat(Literal("<-"), NonTerm("Spacing")),
		"SLASH":     Concat(Literal("/"), NonTerm("Spacing")),
		"AND":       Concat(Literal("&"), NonTerm("Spacing")),
		"NOT":       Concat(Literal("!"), NonTerm("Spacing")),
		"QUESTION":  Concat(Literal("?"), NonTerm("Spacing")),
		"STAR":      Concat(Literal("*"), NonTerm("Spacing")),
		"PLUS":      Concat(Literal("+"), NonTerm("Spacing")),
		"OPEN":      Concat(Literal("("), NonTerm("Spacing")),
		"CLOSE":     Concat(Literal(")"), NonTerm("Spacing")),
		"DOT":       Concat(Literal("."), NonTerm("Spacing")),

		"Spacing":   Star(Or(NonTerm("Space"), NonTerm("Comment"))),
		"Comment":   Concat(Literal("#"), Star(Concat(Not(NonTerm("EndOfLine")), Any(1))), NonTerm("EndOfLine")),
		"Space":     Or(Literal(" "), Literal("\t"), NonTerm("EndOfLine")),
		"EndOfLine": Or(Literal("\r\n"), Literal("\n"), Literal("\r")),
		"EndOfFile": Not(Any(1)),
	}, ids)

	p.Optimize()
	fmt.Println(p)
	fmt.Println(pattern.Histogram(p))

	code := vm.Encode(p)
	fmt.Println(code)
	fmt.Println("Code size", code.Size())

	data, err := ioutil.ReadFile("../testdata/peg.peg")
	if err != nil {
		log.Fatal(err)
	}
	in := input.ByteReader(input.ByteReader(data))
	machine := vm.NewVM(in, code)
	tbl := memo.NoneTable{}

	PrintMemUsage()
	fmt.Printf("Input is %d bytes\n", len(data))

	start := time.Now()
	match, length, caps := machine.Exec(tbl)
	fmt.Println(match, length)
	elapsed := time.Since(start)
	elapsed = time.Since(start)
	fmt.Println("Parse completed in", elapsed)

	root := machine.CaptureAST(caps, code)
	fmt.Println(len(root[0].Children))
}
