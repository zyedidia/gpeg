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

// Slice returns the corresponding slice from the byte reader.
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

// Slice returns the corresponding slice from the string reader.
func (b StringReader) Slice(low, high Pos) []byte {
	return b[low:high]
}
