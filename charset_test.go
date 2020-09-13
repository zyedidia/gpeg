package gpeg

import "testing"

func inSet(set charset, in, notin []rune, t *testing.T) {
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
	in := []rune{'a', 'b', 'c', 'd', '{', '}', '😀', 'τ', 'ι', 'α'}
	notin := []rune{'x', 'y', 'z', '[', ']', 'γ', 'ν', 'ω', 'ρ'}

	set := Charset(in)

	inSet(set, in, notin, t)
}

func TestCharsetRange(t *testing.T) {
	set := CharsetRange('a', 'z').Add(CharsetRange('A', 'Z'))

	in := []rune{'a', 'b', 'c', 'd', 'z', 'y', 'A', 'Z', 'B'}
	notin := []rune{'0', '1', '2', 0}

	inSet(set, in, notin, t)
}

func TestCharsetComplement(t *testing.T) {
	in := []rune{'a', 'b', 'c', 'd', '{', '}', '😀', 'τ', 'ι', 'α'}
	notin := []rune{'x', 'y', 'z', '[', ']', 'γ', 'ν', 'ω', 'ρ'}

	set := Charset(in).Complement()

	inSet(set, notin, in, t)
}
