// Package input defines data types and functions for managing input data.
package input

import (
	"io"
)

const bufsz = 4096

// Input represents the input data and is an efficient wrapper of io.ReaderAt
// which provides a nicer API, avoids repeated interface function calls, and
// uses a cache for buffered reading.
// An Input also tracks the index of the furthest byte that has been read.
type Input struct {
	r io.ReaderAt

	// cached data.
	chunk [bufsz]byte
	b     [1]byte
	// size of the cache.
	nchunk int

	// the position within the reader that the chunk starts at.
	base int
	// the offset within the chunk we are reading at.
	coff int
	// the furthest position we have read.
	furthest int
}

// NewInput creates a new Input wrapper for the io.ReaderAt.
func NewInput(r io.ReaderAt) *Input {
	i := &Input{
		r: r,
	}
	i.refill(i.base)
	return i
}

func (i *Input) refill(pos int) {
	i.base = pos
	i.coff = 0
	i.nchunk, _ = i.r.ReadAt(i.chunk[:], int64(i.base))
}

// Peek returns the next byte in the stream or 'false' if there are no more
// bytes. Successive calls to Peek will return the same value unless there is a
// call to SeekTo or Advance in between.
func (i *Input) Peek() (byte, bool) {
	pos := i.base + i.coff
	if pos > i.furthest {
		i.furthest = pos
	}

	return i.chunk[i.coff], i.nchunk != 0
}

func (i *Input) PeekBefore() (byte, bool) {
	if i.base+i.coff-1 < 0 {
		return 0, false
	}
	if i.coff >= 1 {
		return i.chunk[i.coff-1], i.nchunk != 0
	}
	n, _ := i.r.ReadAt(i.b[:], int64(i.base+i.coff-1))
	return i.b[0], n == 1
}

// SeekTo moves the current read position to the desired read position. Returns
// true if the seek went to a valid location within the reader, and false
// otherwise. In other words, if seek returns true the next call to Peek will
// return a valid byte.
func (i *Input) SeekTo(pos int) bool {
	// check if the seek position in within the current chunk and if so just
	// update the internal offset.
	chunkEnd := i.base + i.nchunk
	if pos < chunkEnd && pos >= i.base {
		i.coff = pos - i.base
		return true
	}

	// refill the cache (moves the base)
	i.refill(pos)
	return i.nchunk != 0
}

// Advance moves the offset forward by 'n' bytes. Returns true if the advance
// was successful (n chars were successfully skipped) and false otherwise. Note
// that even if Advance returns true the next call to Peek may return false if
// the advance went to the exact end of the data.
func (i *Input) Advance(n int) bool {
	if i.nchunk == 0 {
		return false
	}

	i.coff += n
	if i.coff > i.nchunk {
		i.refill(i.base + i.coff)
		return false
	} else if i.coff == i.nchunk {
		i.refill(i.base + i.coff)
	}
	return true
}

// Slice returns a slice of the reader corresponding to the range [low:high).
func (i *Input) Slice(low, high int) []byte {
	return Slice(i.r, low, high)
}

// Pos returns the current read position.
func (i *Input) Pos() int {
	return i.base + i.coff
}

// Furthest returns the furthest read position.
func (i *Input) Furthest() int {
	return i.furthest
}

// ResetFurthest resets the furthest read tracker to zero.
func (i *Input) ResetFurthest() {
	i.furthest = 0
}
