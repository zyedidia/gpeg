package input

import "io"

// A ReaderAtPos is an interface for reading bytes in chunks from a document
// that may have a more complex position representation.
type ReaderAtPos interface {
	// ReadAtPos reads as many bytes as possible from p and returns the
	// result as a slice. It is expected that the data being read is already
	// in memory, and as a result the slice that is returned does not cause
	// any allocation.
	ReadAtPos(p Pos) (b []byte, err error)
}

// A ByteReader implements the input.ReaderAtPos interface for byte slices
type ByteReader []byte

// ReadAtPos returns as many bytes starting at pos as possible.
func (r ByteReader) ReadAtPos(pos Pos) ([]byte, error) {
	if pos.Off >= len(r) {
		return []byte{}, io.EOF
	}
	return r[pos.Off:], nil
}

// A StringReader implements the input.Reader interface for strings
type StringReader string

// ReadAtPos returns as many bytes starting at pos as possible.
func (r StringReader) ReadAtPos(pos Pos) ([]byte, error) {
	if pos.Off >= len(r) {
		return []byte{}, io.EOF
	}
	return []byte(r[pos.Off:]), nil
}
