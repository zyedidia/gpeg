package isa

// Insn represents the interface for an instruction in the ISA
type Insn interface {
	insn()
}

var uniqId int

// Label is used for marking a location in the instruction code with
// a unique ID
type Label struct {
	basic
	Id int
}

// NewLabel returns a new label with a unique ID
func NewLabel() Label {
	uniqId++
	return Label{
		Id: uniqId,
	}
}

// Char consumes the next byte of the subject if it matches `Byte` and
// fails otherwise.
type Char struct {
	basic
	Byte byte
}

// Jump jumps to `Lbl`.
type Jump struct {
	basic
	Lbl Label
}

// Choice pushes `Lbl` to the stack and if there is a failure the label will
// be popped from the stack and jumped to.
type Choice struct {
	basic
	Lbl Label
}

// Call pushes the next instruction to the stack as a return address and jumps
// to `Lbl`.
type Call struct {
	basic
	Lbl Label
}

// Commit jumps to `Lbl` and removes the top entry from the stack
type Commit struct {
	basic
	Lbl Label
}

// Return pops a return address off the stack and jumps to it.
type Return struct {
	basic
}

// Fail causes the instruction pointer to go to the fail state.
type Fail struct {
	basic
}

// Set consumes the next byte of input if it is in the set of chars defined
// by `Chars`.
type Set struct {
	basic
	Chars Charset
}

// Any consumes the next `N` UTF-8 codepoints and fails if that is not
// possible.
type Any struct {
	basic
	N int
}

type PartialCommit struct {
	basic
	Lbl Label
}

// Span consumes zero or more bytes in the set `Chars`. This instruction
// never fails.
type Span struct {
	basic
	Chars Charset
}

type BackCommit struct {
	basic
	Lbl Label
}

type FailTwice struct {
	basic
}

// TestChar consumes the next byte if it matches `Byte` and jumps to `Lbl`
// otherwise.
type TestChar struct {
	basic
	Byte byte
	Lbl  Label
}

// TestSet consumes the next byte if it is in the set `Chars` and jumps to
// `Lbl` otherwise.
type TestSet struct {
	basic
	Chars Charset
	Lbl   Label
}

// TestAny consumes the next `N` UTF-8 codepoints and jumps to `Lbl` if that
// is not possible.
type TestAny struct {
	basic
	N   int
	Lbl Label
}

type Choice2 struct {
	basic
	Lbl  Label
	Back int
}

type basic struct{}

func (b basic) insn() {}
