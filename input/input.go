// Package input defines data types and functions for managing input data.
package input

import "bytes"

var dummy = []byte{0}

// Input represents the input data and is an efficient wrapper of ReaderAtPos
// which provides a nicer interface, avoids repeated interface function calls,
// and uses a cache for buffered reading.
type Input struct {
	r ReaderAtPos

	// the cached data
	chunk []byte
	// the position within 'r' that the chunk corresponds to.
	base Pos
	// the offset within the chunk we are reading at.
	off int
	// the furthest position we have read.
	furthest Pos
	// true if the offset has reached the end of the reader
	finished bool
}

// NewInput creates a new Input wrapper for the ReaderAtPos.
func NewInput(r ReaderAtPos) *Input {
	i := &Input{
		r: r,
	}
	i.refill(i.base)
	return i
}

func (i *Input) refill(pos Pos) {
	var err error

	i.base = pos
	i.off = 0
	i.chunk, err = i.r.ReadAtPos(i.base)
	// reached the end of the reader
	i.finished = err != nil || len(i.chunk) == 0
	if i.finished {
		// set the chunk to a dummy for peeks to read from
		i.chunk = dummy
	}
}

// Peek returns the next byte in the stream or 'false' if there are no more
// bytes. Successive calls to Peek will return the same value unless there is a
// call to SeekTo or Advance in between.
func (i *Input) Peek() (byte, bool) {
	pos := i.base.Move(i.off)
	if pos.Cmp(i.furthest) > 0 {
		i.furthest = pos
	}

	return i.chunk[i.off], !i.finished
}

// SeekTo moves the current read position to the desired read position. Returns
// true if the seek went to a valid location within the reader, and false
// otherwise.
func (i *Input) SeekTo(pos Pos) bool {
	// check if the seek position is within the current chunk and if so just
	// update the internal offset.
	chunkEnd := i.base.Move(len(i.chunk))
	off := pos.Cmp(i.base)
	if pos.Cmp(chunkEnd) < 0 && pos.Cmp(i.base) >= 0 {
		i.off = off
		return true
	}

	// refill the cache (moves the base)
	i.refill(pos)
	return !i.finished
}

// Advance moves the offset forward by 'n' bytes. Returns true if the advance
// was successful (n chars were successfully skipped), and false otherwise.
func (i *Input) Advance(n int) bool {
	if i.finished {
		return false
	}

	i.off += n
	if i.off > len(i.chunk) {
		// moved past the end
		i.refill(i.base.Move(i.off))
		return false
	} else if i.off == len(i.chunk) {
		// moved exactly to the end
		i.refill(i.base.Move(i.off))
	}
	return true
}

// Slice returns a slice of the reader corresponding to the range [low:high).
func (i *Input) Slice(low, high Pos) []byte {
	buf := &bytes.Buffer{}
	off := low
	for nleft := high.Cmp(off); nleft > 0; nleft = high.Cmp(off) {
		b, err := i.r.ReadAtPos(off)
		if err != nil {
			break
		}
		sz := min(len(b), nleft)
		n, _ := buf.Write(b[:sz])
		off = off.Move(n)
	}
	return buf.Bytes()
}

// Pos returns the current read position.
func (i *Input) Pos() Pos {
	return i.base.Move(i.off)
}

// Furthest returns the furthest read position.
func (i *Input) Furthest() Pos {
	return i.furthest
}

// ResetFurthest resets the furthest read position to its zero value.
func (i *Input) ResetFurthest() {
	var p Pos
	i.furthest = p
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
