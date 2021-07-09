package memo

import (
	"github.com/zyedidia/gpeg/memo/interval"
)

// An Entry represents a memoized parse result. It stores the non-terminal
// memoized, the start position of the parse result, the length, and the number
// of characters examined to make the parse determination. If the length is -1,
// the non-terminal failed to match at this location (but still may have
// examined a non-zero number of characters).
type Entry struct {
	length   int
	examined int
	count    int
	captures []*Capture
	pos      interval.Pos
}

func (e *Entry) setPos(pos interval.Pos) {
	e.pos = pos
	for i := range e.captures {
		e.captures[i].setMEnt(e)
	}
}

// Pos returns this entry's starting position.
func (e *Entry) Pos() int {
	return e.pos.Pos()
}

// Length returns the number of characters memoized by this entry.
func (e *Entry) Length() int {
	return e.length
}

// Captures returns the captures that occurred within this memoized parse
// result.
func (e *Entry) Captures() []*Capture {
	return e.captures
}

func (e *Entry) Count() int {
	return e.count
}

func (e *Entry) Examined() int {
	return e.examined
}
