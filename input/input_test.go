package input_test

import (
	"bytes"
	"testing"

	"github.com/zyedidia/gpeg/input"
)

func TestInput(t *testing.T) {
	b := bytes.NewReader([]byte("foo bar baz"))
	i := input.NewInput(b)

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

	slice := i.Slice(4, 7)
	if string(slice) != "bar" {
		t.Error("incorrect slice, got", string(slice))
	}

	success := i.Advance(9)
	if !success {
		t.Error("incorrect: couldn't advance by 9")
	}

	if b, ok := i.Peek(); ok {
		t.Errorf("peek past end of buffer should return false, got %c", b)
	}
}
