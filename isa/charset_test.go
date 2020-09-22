package isa

import (
	"testing"
)

func inSet(set Charset, in, notin []byte, t *testing.T) {
	for _, r := range in {
		if !set.Has(r) {
			t.Errorf("Error: %c returned 'not in set'", r)
		}
	}

	for _, r := range notin {
		if set.Has(r) {
			t.Errorf("Error: %c returned 'in set'", r)
		}
	}
}

func TestCharset(t *testing.T) {
	in := []byte{'a', 'b', 'c', 'd', '{', '}'}
	notin := []byte{'x', 'y', 'z', '[', ']'}

	set := NewCharset(in)

	inSet(set, in, notin, t)
}

func TestCharsetRangeUnion(t *testing.T) {
	set := CharsetRange('a', 'z').Add(CharsetRange('A', 'Z'))

	in := []byte{'a', 'b', 'c', 'd', 'z', 'y', 'A', 'Z', 'B'}
	notin := []byte{'0', '1', '2', 0}

	inSet(set, in, notin, t)
}

func TestCharsetComplement(t *testing.T) {
	in := []byte{'a', 'b', 'c', 'd', '{', '}'}
	notin := []byte{'x', 'y', 'z', '[', ']'}

	set := NewCharset(in).Complement()

	inSet(set, notin, in, t)
}
