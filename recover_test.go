package gpeg

import (
	"strings"
	"testing"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func sync(p Pattern) Pattern {
	return Star(Concat(Not(p), Any(1)))
}

func TestRecover(t *testing.T) {
	id := Plus(Set(charset.Range('a', 'z')))
	p := Grammar("S", map[string]Pattern{
		"S": Or(NonTerm("List"), Concat(Any(1), Error("expecting a list of identifiers", NonTerm("ErrList")))),
		"List": Concat(
			NonTerm("Id"),
			Star(Concat(And(Any(1)),
				NonTerm("Comma"),
				Or(NonTerm("Id"),
					Error("expecting an identifier", NonTerm("ErrId")))),
			),
		),
		"Id":       Concat(NonTerm("Sp"), Cap(id)),
		"Comma":    Or(Concat(NonTerm("Sp"), Literal(",")), Error("expecting ','", NonTerm("ErrComma"))),
		"Sp":       Star(Set(charset.New([]byte{' ', '\n', '\t'}))),
		"ErrId":    sync(Literal(",")),
		"ErrComma": sync(id),
		"ErrList":  sync(Not(Any(1))),
	})

	peg := MustCompile(p)
	code := vm.Encode(peg)
	in := strings.NewReader("one two three,")
	_, _, _, errs := code.Exec(in, memo.NoneTable{})

	if len(errs) != 3 {
		t.Error("Incorrect list of errors:", errs)
	}
}
