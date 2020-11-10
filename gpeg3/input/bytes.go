package input

import (
	"io"
)

// A ByteReader implements the input.Reader interface for byte slices
type ByteReader []byte

// ReadAtPos returns as many bytes starting at pos as possible.
func (b ByteReader) ReadAtPos(pos Pos) ([]byte, error) {
	if int(pos) >= len(b) {
		return []byte{}, io.EOF
	}
	return b[pos:], nil
}

func (b ByteReader) Slice(low, high Pos) []byte {
	return b[low:high]
}

// A StringReader implements the input.Reader interface for strings
type StringReader []byte

// ReadAtPos returns as many bytes starting at pos as possible.
func (b StringReader) ReadAtPos(pos Pos) ([]byte, error) {
	if int(pos) >= len(b) {
		return []byte{}, io.EOF
	}
	return b[pos:], nil
}

func (b StringReader) Slice(low, high Pos) []byte {
	if int(high) > len(b) {
		return []byte{}
	}
	return b[low:high]
}
