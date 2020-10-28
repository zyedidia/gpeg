package vm

import (
	"testing"

	"github.com/zyedidia/gpeg/charset"
	. "github.com/zyedidia/gpeg/pattern"
)

func TestCode(t *testing.T) {
	p := Grammar("Expr", map[string]Pattern{
		"Expr":   Concat(NonTerm("Factor"), Star(Concat(Set(charset.New([]byte{'+', '-'})), NonTerm("Factor")))),
		"Factor": Concat(NonTerm("Term"), Star(Concat(Set(charset.New([]byte{'*', '/'})), NonTerm("Term")))),
		"Term":   Or(NonTerm("Number"), Concat(Concat(Literal("("), NonTerm("Expr")), Literal(")"))),
		"Number": Plus(Set(charset.Range('0', '9'))),
	})

	code := Encode(p)
	b := code.Bytes()
	load := LoadCode(b)

	if load.Size() != code.Size() {
		t.Error("Saved and loaded code not equivalent")
	}

	for i := range code.data.Insns {
		if load.data.Insns[i] != code.data.Insns[i] {
			t.Errorf("Code byte %d does not match", i)
		}
	}
}
