package input

import "io"

// BytePos implements the Pos interface using a simple byte offset.
type BytePos int

const BytePosStart = BytePos(0)

// Distance returns the number of bytes between b and p (always positive).
func (b BytePos) Distance(p Pos) int {
	if b > p.(BytePos) {
		return int(b - p.(BytePos))
	}
	return int(p.(BytePos) - b)
}

// Advance increases this position by n bytes.
func (b BytePos) Advance(n int) Pos {
	return BytePos(int(b) + n)
}

// A ByteReader implements the input.Reader interface for byte slices
type ByteReader []byte

// ReadAtPos reads up to len(p) bytes into p starting at Pos. Pos must be a
// BytePos.
func (b ByteReader) ReadAtPos(pos Pos) ([]byte, error) {
	bp := pos.(BytePos)
	if int(bp) >= len(b) {
		return []byte{}, io.EOF
	}
	return b[bp:], nil
}
