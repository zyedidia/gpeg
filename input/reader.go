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

func (b ByteReader) Slice(low, high Pos) []byte {
	return b[low:high]
}
