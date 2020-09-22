package input

import (
	"io"
)

// A ByteReader implements the input.Reader interface for byte slices
type ByteReader []byte

// ReadAtPos reads up to len(p) bytes into p starting at Pos.
func (b ByteReader) ReadAtPos(pos Pos) ([]byte, error) {
	p := pos.(uint32)
	if int(p) >= len(b) {
		return []byte{}, io.EOF
	}
	return b[p:], nil
}

func (b ByteReader) Advance(p Pos, n int) Pos {
	return p.(uint32) + uint32(n)
}

// A ByteReader implements the input.Reader interface for strings
type StringReader string

// ReadAtPos reads up to len(p) bytes into p starting at Pos.
func (s StringReader) ReadAtPos(pos Pos) ([]byte, error) {
	p := pos.(uint32)
	if int(p) >= len(s) {
		return []byte{}, io.EOF
	}
	return []byte(s[p:]), nil
}
