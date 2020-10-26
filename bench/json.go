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

func json() {
	p := Grammar("doc", map[string]Pattern{
		"Space":         Set(charset.New([]byte{9, 10, 11, 12, 13, ' '})),
		"Xdigit":        Set(charset.Range('A', 'F').Add(charset.Range('a', 'f')).Add(charset.Range('0', '9'))),
		"Digit":         Set(charset.Range('0', '9')),
		"S":             Star(NonTerm("Space")),
		"jtrue":         Literal("true"),
		"jfalse":        Literal("false"),
		"jnull":         Literal("null"),
		"unicodeEscape": Concat(Literal("u"), Repeat(NonTerm("Xdigit"), 4)),
		"escape":        Concat(Literal("\\"), Or(Set(charset.New([]byte{'{', '"', '|', '\\', 'b', 'f', 'n', 'r', 't'})), NonTerm("unicodeEscape"))),
		"stringBody":    Concat(Optional(NonTerm("escape")), Star(Concat(Plus(Set(charset.Range('\x20', '\xff').Sub(charset.New([]byte{'"', '\\'})))), Star(NonTerm("escape"))))),
		"jstring":       Concat(Optional(NonTerm("S")), Literal("\""), NonTerm("stringBody"), Literal("\""), Optional(NonTerm("S"))),
		"minus":         Literal("-"),
		"intPart":       Or(Literal("0"), Concat(Concat(Not(Literal("0")), NonTerm("Digit")), Star(NonTerm("Digit")))),
		"fractPart":     Concat(Literal("."), Plus(NonTerm("Digit"))),
		"expPart":       Concat(Or(Literal("e"), Literal("E")), Optional(Or(Literal("+"), Literal("-"))), Plus(NonTerm("Digit"))),
		"jnumber":       Concat(Optional(NonTerm("minus")), NonTerm("intPart"), Optional(NonTerm("fractPart")), Optional(NonTerm("expPart"))),
		"doc":           Concat(NonTerm("JSON"), Not(Any(1))),
		"JSON":          Memo(Concat(Optional(NonTerm("S")), Or(NonTerm("jnumber"), NonTerm("jobject"), NonTerm("jarray"), NonTerm("jstring"), NonTerm("jtrue"), NonTerm("jfalse"), NonTerm("jnull")), Optional(NonTerm("S")))),
		"jobject":       Concat(Literal("{"), Or(Concat(NonTerm("jstring"), Literal(":"), NonTerm("JSON"), Star(Concat(Literal(","), NonTerm("jstring"), Literal(":"), NonTerm("JSON")))), Optional(NonTerm("S"))), Literal("}")),
		"jarray":        Concat(Literal("["), Or(Concat(NonTerm("JSON"), Star(Concat(Literal(","), NonTerm("JSON")))), Optional(NonTerm("S"))), Literal("]")),
	})

	p.Optimize()
	fmt.Println(p)

	code := vm.Encode(p)
	fmt.Println(code)

	data, err := ioutil.ReadFile("../testdata/test.json")
	if err != nil {
		log.Fatal(err)
	}
	in := input.ByteReader(input.ByteReader(data))
	// in := input.ByteReader(json)
	machine := vm.NewVM(in, code)
	tbl := memo.NewMapTable()

	PrintMemUsage()

	start := time.Now()
	fmt.Println(machine.Exec(tbl))
	elapsed := time.Since(start)
	elapsed = time.Since(start)
	fmt.Println("Parse completed in", elapsed)

	fmt.Printf("Memo table has %d entries\n", tbl.Size())
	start = time.Now()
	tbl.ApplyEdit(memo.Edit{
		Start: 150,
		End:   151,
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
