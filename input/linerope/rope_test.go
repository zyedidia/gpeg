package linerope_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/zyedidia/gpeg/input/linerope"
)

func check(r *linerope.Node, b *basicText, t *testing.T) {
	if !bytes.Equal(r.Value(), b.value()) {
		t.Errorf("incorrect bytes: %s %s", string(r.Value()), string(b.value()))
	}
	if r.Len() != b.length() {
		t.Errorf("incorrect length: %d %d", r.Len(), b.length())
	}
	if r.NumLines() != b.NumLines() {
		t.Errorf("incorrect line count: %d %d", r.NumLines(), b.NumLines())
	}

	const ncheck = 100
	for i := 0; i < ncheck; i++ {
		pos := rand.Intn(r.Len())
		rline, rcol := r.LineColAt(pos)
		bline, bcol := b.lineColAt(pos)
		if rline != bline || rcol != bcol {
			t.Errorf("incorrect offset conversion: %d, want (%d, %d), got (%d, %d)", pos, bline, bcol, rline, rcol)
		}

		off := r.OffsetAt(rline, rcol)
		if off != pos {
			t.Errorf("incorrect line/col conversion: (%d, %d), want %d, got %d", rline, rcol, pos, off)
		}
	}
}

const datasz = 5000

func data() (*linerope.Node, *basicText) {
	data := randbytes(datasz)
	r := linerope.New(data)
	b := newBasicText(data)
	return r, b
}

func randrange(high int) (int, int) {
	i1 := rand.Intn(high)
	i2 := rand.Intn(high)
	return min(i1, i2), max(i1, i2)
}

var letters = []byte("\nabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randbytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func TestConstruction(t *testing.T) {
	r, b := data()
	check(r, b, t)
}

func TestInsertRemove(t *testing.T) {
	r, b := data()

	const nedit = 100
	const strlen = 20
	for i := 0; i < nedit; i++ {
		low, high := randrange(r.Len())
		r.Remove(low, high)
		b.remove(low, high)
		check(r, b, t)
		bstr := randbytes(strlen)
		r.Insert(low, bstr)
		b.insert(low, bstr)
		check(r, b, t)
	}
	check(r, b, t)
}

func TestReadAt(t *testing.T) {
	r, b := data()

	const nslice = 100
	length := r.Len()
	for i := 0; i < nslice; i++ {
		low, high := randrange(length)

		rb := make([]byte, high-low)
		r.ReadAt(rb, int64(low))
		bb := b.slice(low, high)
		if !bytes.Equal(rb, bb) {
			t.Errorf("slice not equal: %s %s", string(rb), string(bb))
		}
	}
}

func TestSplit(t *testing.T) {
	r, b := data()

	const nsplit = 10
	for i := 0; i < nsplit; i++ {
		splitidx := rand.Intn(r.Len())
		left, right := r.SplitAt(splitidx)

		lb := b.slice(0, splitidx)
		rb := b.slice(splitidx, b.length())
		if !bytes.Equal(left.Value(), lb) {
			t.Errorf("%d: left slice not equal: %s %s", splitidx, string(left.Value()), string(lb))
		}
		if !bytes.Equal(right.Value(), rb) {
			t.Errorf("%d: right slice not equal: %s %s", splitidx, string(right.Value()), string(rb))
		}
		r = linerope.Join(left, right)
		check(r, b, t)
	}
}

type basicText struct {
	data []byte
}

func newBasicText(b []byte) *basicText {
	data := make([]byte, len(b))
	copy(data, b)
	return &basicText{
		data: data,
	}
}

func (b *basicText) length() int {
	return len(b.data)
}

func (b *basicText) value() []byte {
	return b.data
}

func (b *basicText) remove(start, end int) {
	b.data = append(b.data[:start], b.data[end:]...)
}

func (b *basicText) insert(pos int, val []byte) {
	b.data = insert(b.data, pos, val)
}

func (b *basicText) slice(start, end int) []byte {
	return b.data[start:end]
}

func (b *basicText) lineColAt(pos int) (line, col int) {
	var last int
	for i, c := range b.data {
		if c == '\n' {
			if i >= pos {
				return line, pos - last
			}
			last = i + 1
			line++
		}
	}
	return line, pos - last
}

func (b *basicText) NumLines() int {
	return bytes.Count(b.data, []byte{'\n'})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// from slice tricks
func insert(s []byte, k int, vs []byte) []byte {
	if n := len(s) + len(vs); n <= cap(s) {
		s2 := s[:n]
		copy(s2[k+len(vs):], s[k:])
		copy(s2[k:], vs)
		return s2
	}
	s2 := make([]byte, len(s)+len(vs))
	copy(s2, s[:k])
	copy(s2[k:], vs)
	copy(s2[k+len(vs):], s[k:])
	return s2
}
