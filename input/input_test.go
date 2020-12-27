package input

import "testing"

func TestInput(t *testing.T) {
	s := StringReader("foo bar baz")
	i := NewInput(s)

	if b, _ := i.Peek(); b != 'f' {
		t.Error("incorrect peek, got", string(b))
	}
	i.Advance(1)
	if b, _ := i.Peek(); b != 'o' {
		t.Error("incorrect peek, got", string(b))
	}
	i.Advance(1)
	if b, _ := i.Peek(); b != 'o' {
		t.Error("incorrect peek, got", string(b))
	}

	slice := i.Slice(Pos{4}, Pos{7})
	if string(slice) != "bar" {
		t.Error("incorrect slice, got", string(slice))
	}
}
