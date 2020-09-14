package gpeg

import (
	"unicode/utf8"
)

const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

// A Reader defines the interface needed for a pattern to match some input.
type Reader interface {
	// PeekRune returns the next rune in the reader without advancing
	// and the number of bytes reader would advance if that rune was read.
	PeekRune() (rune, int)
	// SeekBytes seeks the reader to a new byte offset.
	SeekBytes(off, whence int)
	// Offset returns the current byte offset of the reader.
	Offset() int
	// Len returns the number of bytes in the reader.
	Len() int
}

// StringReader implements the reader interface for strings.
type StringReader struct {
	data string
	pos  int
}

func NewStringReader(s string) *StringReader {
	return &StringReader{
		data: s,
		pos:  0,
	}
}

func (s *StringReader) PeekRune() (rune, int) {
	return utf8.DecodeRuneInString(s.data[s.pos:])
}

func (s *StringReader) SeekBytes(off, whence int) {
	switch whence {
	case SeekStart:
		s.pos = off
	case SeekCurrent:
		s.pos += off
	case SeekEnd:
		s.pos = len(s.data) - off
	}
}

func (s *StringReader) Offset() int {
	return s.pos
}

func (s *StringReader) Len() int {
	return len(s.data)
}

// ByteReader implements the Reader interface for byte slices.
type ByteReader struct {
	data []byte
	pos  int
}

func NewByteReader(b []byte) *ByteReader {
	return &ByteReader{
		data: b,
		pos:  0,
	}
}

func (s *ByteReader) PeekRune() (rune, int) {
	return utf8.DecodeRune(s.data[s.pos:])
}

func (s *ByteReader) SeekBytes(off, whence int) {
	switch whence {
	case SeekStart:
		s.pos = off
	case SeekCurrent:
		s.pos += off
	case SeekEnd:
		s.pos = len(s.data) - off
	}
}

func (s *ByteReader) Offset() int {
	return s.pos
}

func (s *ByteReader) Len() int {
	return len(s.data)
}
