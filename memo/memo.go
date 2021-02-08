// Package memo defines data structures and functions for creating memoization
// tables.
package memo

import (
	"fmt"

	"github.com/zyedidia/gpeg/capture"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo/avlint"
)

// A Key is used to look up memo.Entry values. It stores an Id representing
// the pattern being looked up, and the position in the subject string to
// look up the entry for. For example, if we are memoizing the non-terminal
// "Number", the key would be Key{id("Number"), textpos} to check if a number
// has already been parsed and memoized at textpos.
type Key struct {
	Id  int16
	Pos input.Pos
}

// An Entry represents memoization information in the memo.Table. The entry
// stores the number of characters examined to parse the pattern being
// memoized, and the length of the match.
type Entry struct {
	id       int16
	start    input.Pos
	examined int
	length   int
	val      []*capture.Node
}

// NewEntry returns a new entry with the given information.
func NewEntry(id int16, start input.Pos, matchlen, examlen int, val []*capture.Node) *Entry {
	e := &Entry{
		id:       id,
		start:    start,
		examined: examlen,
		length:   matchlen,
		val:      val,
	}
	for _, c := range e.val {
		c.Loc = e
	}
	return e
}

// MatchLength returns the match length of this entry
func (e *Entry) MatchLength() int {
	return e.length
}

// MaxExamined returns the number of characters that were examined to parse
// this entry's pattern.
func (e *Entry) Examined() int {
	return e.examined
}

// Value returns the parse result associated with this memo entry.
func (e *Entry) Value() []*capture.Node {
	return e.val
}

// Start returns the start position of this memo entry.
func (e *Entry) Start() input.Pos {
	return e.start
}

// End returns the end position of this memo entry.
func (e *Entry) End() input.Pos {
	return e.start.Move(e.length)
}

func (e *Entry) Low() int {
	return e.start.Off
}

func (e *Entry) ShiftLow(count int) {
	e.start.Move(count)
}

func (e *Entry) High() int {
	return e.start.Off + e.examined
}

func (e *Entry) ShiftHigh(count int) {
	// t.end += count
}

func (e *Entry) Overlaps(i avlint.Interval) bool {
	return e.Low() < i.High() && e.High() > i.Low()
}

func (e *Entry) String() string {
	return fmt.Sprintf("[%d, %d)", e.Low(), e.High())
}
