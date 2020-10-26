package charset

import (
	"testing"
)

func inSet(set Set, in, notin []byte, t *testing.T) {
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

func TestSet(t *testing.T) {
	in := []byte{'a', 'b', 'c', 'd', '{', '}'}
	notin := []byte{'x', 'y', 'z', '[', ']'}

	set := New(in)

	inSet(set, in, notin, t)
}

func TestRangeUnion(t *testing.T) {
	set := Range('a', 'z').Add(Range('A', 'Z'))

	in := []byte{'a', 'b', 'c', 'd', 'z', 'y', 'A', 'Z', 'B'}
	notin := []byte{'0', '1', '2', 0}

	inSet(set, in, notin, t)
}

func TestComplement(t *testing.T) {
	in := []byte{'a', 'b', 'c', 'd', '{', '}'}
	notin := []byte{'x', 'y', 'z', '[', ']'}

	set := New(in).Complement()

	inSet(set, notin, in, t)
}

func TestBigSet(t *testing.T) {
	in := []byte{200, 201, 203}
	notin := []byte{0, 1, 2}

	set := Range(128, '\xff')

	inSet(set, in, notin, t)
}
