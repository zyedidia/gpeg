package gpeg

import (
	"unicode/utf8"
)

const (
	SeekStart   = 0 // seek relative to the origin of the file
	SeekCurrent = 1 // seek relative to the current offset
	SeekEnd     = 2 // seek relative to the end
)

type Reader interface {
	PeekRune() (rune, int)
	SeekBytes(off, whence int)
	Offset() int
	Len() int
}

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
	// return rune(s.data[s.pos]), 1
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
