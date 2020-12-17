package memo

import (
	"github.com/zyedidia/gpeg/ast"
	"github.com/zyedidia/gpeg/input"
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
	examined int
	length   int
	val      []*ast.Node
}

// NewEntry returns a new entry with the given information.
func NewEntry(matchlen, examlen int, val []*ast.Node) Entry {
	return Entry{
		examined: examlen,
		length:   matchlen,
		val:      val,
	}
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

func (e *Entry) Value() []*ast.Node {
	return e.val
}
