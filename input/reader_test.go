package input_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/zyedidia/gpeg/input"
)

func TestReaderWrapper(t *testing.T) {
	r := bytes.NewBufferString("foo bar baz")
	rat := input.FromReader(r)
	b := make([]byte, 3)

	rat.ReadAt(b, 4)
	if string(b) != "bar" {
		t.Errorf("want %s, got %s", "bar", string(b))
	}

	rat.ReadAt(b, 0)
	if string(b) != "foo" {
		t.Errorf("want %s, got %s", "foo", string(b))
	}

	n, err := rat.ReadAt(b, 9)
	if string(b[:n]) != "az" {
		t.Errorf("want %s, got %s", "az", string(b))
	}
	if n != 2 || err != io.EOF {
		t.Errorf("incorrect, n: %v, err: %v", n, err)
	}
}
