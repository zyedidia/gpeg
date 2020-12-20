package capture

import "github.com/zyedidia/gpeg/input"

// A Capture may be any type and is the result from parsing. Usually captures
// will be one of the predefined capture types from this package, but they may
// also be user types that are created by a user-provided capture function.
type Capture interface{}

// A MoveCapture is a capture that requires updating when an edit is made (only
// relevant for incremental parses). This is usually a capture that stores some
// position information.
type MoveCapture interface {
	// Move shifts the position of this capture by the given offset (may be
	// negative.
	Move(off int)
}

// A SubCapture stores the positions in the source document that this capture
// matched, and a string value. Captures of this type may be used for
// substition, where the region of the source is replaced by the captured value
// in the result of the overall substitution.
type SubCapture interface {
	// Start returns the starting position of this capture.
	Start() input.Pos
	// End returns the ending position of this capture.
	End() input.Pos
	// String returns the string value of this capture.
	String() string

	MoveCapture
}

type String struct {
	val string
}

func NewString(id int16, start input.Pos, size int, captures []Capture, in *input.BufferedReader) Capture {
	return &String{
		val: string(in.Slice(start, start+input.Pos(size))),
	}
}

type StringLoc struct {
	val   string
	start input.Pos
	size  int
}

func NewStringLoc(id int16, start input.Pos, size int, captures []Capture, in *input.BufferedReader) Capture {
	return &StringLoc{
		start: start,
		size:  size,
		val:   string(in.Slice(start, start+input.Pos(size))),
	}
}

func (c *StringLoc) Move(size int) {
	c.start += input.Pos(size)
}

type Position struct {
	val input.Pos
}

func (c *Position) Move(size int) {
	c.val += input.Pos(size)
}

type List struct {
	val []Capture
}
