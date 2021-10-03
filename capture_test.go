package gpeg

import (
	"strings"
	"testing"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func TestCaptures(t *testing.T) {
	const (
		digit = iota
		num
	)

	p := Star(Memo(Concat(
		Cap(Plus(
			Cap(Set(charset.Range('0', '9')), digit),
		), num),
		Optional(Literal(" ")),
	)))
	code := vm.Encode(MustCompile(p))
	r := strings.NewReader("12 34 56 78 9")
	_, _, ast, _ := code.Exec(r, memo.NoneTable{})

	expect := [][2]int{
		{0, 2},
		{3, 2},
		{6, 2},
		{9, 2},
		{12, 1},
	}

	it := ast.ChildIterator(0)
	i := 0
	for ch := it(); ch != nil; ch = it() {
		if expect[i][0] != ch.Start() || expect[i][1] != ch.Len() {
			t.Fatal(ch.Start(), ch.Len())
		}
		i++
	}
}
