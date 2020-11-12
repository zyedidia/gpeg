package gpeg

import (
	"testing"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

func TestCaptureIndex(t *testing.T) {
	wordChar := charset.Range('A', 'Z').Add(charset.Range('a', 'z'))
	p := Star(Concat(Star(Set(wordChar.Complement())), Cap(Plus(Set(wordChar)))))
	code := vm.Encode(MustCompile(p))

	var bytes input.ByteReader = []byte("a few more words")
	machine := vm.NewVM(bytes, code)
	_, _, capt := machine.Exec(memo.NoneTable{})
	results := machine.CapturesIndex(capt, code)
	expected := [][2]input.Pos{
		[2]input.Pos{0, 1},
		[2]input.Pos{2, 5},
		[2]input.Pos{6, 10},
		[2]input.Pos{11, 16},
	}

	for i, r := range results {
		if r[0] != expected[i][0] || r[1] != expected[i][1] {
			t.Errorf("Error: got %v", results)
		}
	}
}

func TestCaptureString(t *testing.T) {
	wordChar := charset.Range('A', 'Z').Add(charset.Range('a', 'z'))
	p := Star(Concat(Star(Set(wordChar.Complement())), Cap(Plus(Set(wordChar)))))
	code := vm.Encode(MustCompile(p))

	var bytes input.ByteReader = []byte("a few more words")
	machine := vm.NewVM(bytes, code)
	_, _, capt := machine.Exec(memo.NoneTable{})
	results := machine.CapturesString(capt, code)
	expected := []string{"a", "few", "more", "words"}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("Error: got %v", results)
		}
	}
}
