package input

import (
	"io"
)

// A ByteReader implements the input.Reader interface for byte slices
type ByteReader []byte

// ReadAtPos reads up to len(p) bytes into p starting at Pos.
func (b ByteReader) ReadAtPos(pos Pos) ([]byte, error) {
	if int(pos) >= len(b) {
		return []byte{}, io.EOF
	}
	return b[pos:], nil
}

// A ByteReader implements the input.Reader interface for strings
type StringReader string

// ReadAtPos reads up to len(p) bytes into p starting at Pos.
func (s StringReader) ReadAtPos(pos Pos) ([]byte, error) {
	if int(pos) >= len(s) {
		return []byte{}, io.EOF
	}
	return []byte(s[pos:]), nil
}
