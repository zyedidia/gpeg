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
	_, _, capt, _ := machine.Exec(memo.NoneTable{})
	results := machine.CapturesIndex(capt)
	expected := [][2]input.Pos{
		[2]input.Pos{input.PosFromOff(0), input.PosFromOff(1)},
		[2]input.Pos{input.PosFromOff(2), input.PosFromOff(5)},
		[2]input.Pos{input.PosFromOff(6), input.PosFromOff(10)},
		[2]input.Pos{input.PosFromOff(11), input.PosFromOff(16)},
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
	peg := MustCompile(p)
	code := vm.Encode(peg)

	var bytes input.ByteReader = []byte("a few more words")
	machine := vm.NewVM(bytes, code)
	_, _, capt, _ := machine.Exec(memo.NoneTable{})
	results := machine.CapturesString(capt)
	expected := []string{"a", "few", "more", "words"}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("Error: got %v", results)
		}
	}
}

func TestCaptureBacktrack(t *testing.T) {
	p := Or(Concat(CapId(Literal("abc"), 0), Literal("def")), CapId(Literal("abcdea"), 1))
	peg := MustCompile(p)
	code := vm.Encode(peg)

	var bytes input.ByteReader = []byte("abcdea")
	machine := vm.NewVM(bytes, code)
	_, _, capt, _ := machine.Exec(memo.NoneTable{})

	results := machine.CapturesIndex(capt)
	expected := [][2]input.Pos{
		[2]input.Pos{input.PosFromOff(0), input.PosFromOff(6)},
	}

	for i, r := range results {
		if r[0] != expected[i][0] || r[1] != expected[i][1] {
			t.Errorf("Error: got %v", results)
		}
	}
}
