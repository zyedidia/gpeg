package memo

import (
	"fmt"
)

type locator interface {
	Start() int
}

// An Entry represents a memoized parse result. It stores the non-terminal
// memoized, the start position of the parse result, the length, and the number
// of characters examined to make the parse determination. If the length is -1,
// the non-terminal failed to match at this location (but still may have
// examined a non-zero number of characters).
type Entry struct {
	id       int
	start    int
	length   int
	examined int
	count    int
	captures []*Capture

	loc locator
}

func newEntry(id, start, length, examined, count int, captures []*Capture) *Entry {
	e := &Entry{
		id:       id,
		start:    start,
		length:   length,
		examined: examined,
		captures: captures,
		count:    count,
	}

	for _, c := range captures {
		if !c.memoized() {
			c.setMemo(e)
		}
	}
	return e
}

// Id returns the non-terminal ID of this entry.
func (e *Entry) Id() int {
	return e.id
}

// Start returns the starting offset of this memoized entry.
func (e *Entry) Start() int {
	if e.loc != nil {
		return e.loc.Start()
	}
	return e.start
}

// Length returns the number of characters memoized by this entry.
func (e *Entry) Length() int {
	return e.length
}

// Examined returns the number of characters that were examined to create this
// memo entry.
func (e *Entry) Examined() int {
	return e.examined
}

// Captures returns the captures that occurred within this memoized parse
// result.
func (e *Entry) Captures() []*Capture {
	return e.captures
}

func (e *Entry) Count() int {
	return e.count
}

// String representation of the memo entry.
func (e *Entry) String() string {
	start := e.Start()
	return fmt.Sprintf("(%d) len: [%d, %d), exam: [%d, %d)", e.id, start, start+e.Length(), start, start+e.Examined())
}
