package isa

import "fmt"

// Insn represents the interface for an instruction in the ISA
type Insn interface {
	insn()
}

// A JumpType instruction is any instruction that refers to a Label.
type JumpType interface {
	jumpt()
}

var uniqId int

// Label is used for marking a location in the instruction code with
// a unique ID
type Label struct {
	Id int
	basic
}

// NewLabel returns a new label with a unique ID
func NewLabel() Label {
	uniqId++
	return Label{
		Id: uniqId,
	}
}

// Char consumes the next byte of the subject if it matches Byte and
// fails otherwise.
type Char struct {
	Byte byte
	basic
}

// Jump jumps to Lbl.
type Jump struct {
	Lbl Label
	jump
}

// Choice pushes Lbl to the stack and if there is a failure the label will
// be popped from the stack and jumped to.
type Choice struct {
	Lbl Label
	jump
}

// Call pushes the next instruction to the stack as a return address and jumps
// to Lbl.
type Call struct {
	Lbl Label
	jump
}

// Commit jumps to Lbl and removes the top entry from the stack
type Commit struct {
	Lbl Label
	jump
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
// by Chars.
type Set struct {
	Chars Charset
	basic
}

// Any consumes the next N bytes and fails if that is not possible.
type Any struct {
	N byte
	basic
}

// PartialCommit modifies the backtrack entry on the top of the stack to
// point to the current subject offset, and jumps to Lbl.
type PartialCommit struct {
	Lbl Label
	jump
}

// Span consumes zero or more bytes in the set Chars. This instruction
// never fails.
type Span struct {
	Chars Charset
	basic
}

// BackCommit pops a backtrack entry off the stack, goes to the subject
// position in the entry, and jumps to Lbl.
type BackCommit struct {
	Lbl Label
	jump
}

// FailTwice pops an entry off the stack and sets the instruction pointer to
// the fail state.
type FailTwice struct {
	basic
}

// TestChar consumes the next byte if it matches Byte and jumps to Lbl
// otherwise. If the consumption is possible, a backtrack entry referring
// to Lbl and the subject position from before consumption is pushed to the
// stack.
type TestChar struct {
	Byte byte
	Lbl  Label
	jump
}

// TestSet consumes the next byte if it is in the set Chars and jumps to
// Lbl otherwise. If the consumption is possible, a backtrack entry referring
// to Lbl and the subject position from before consumption is pushed to the
// stack.
type TestSet struct {
	Chars Charset
	Lbl   Label
	jump
}

// TestAny consumes the next N bytes and jumps to Lbl if that is not possible.
// If the consumption is possible, a backtrack entry referring to Lbl and
// the subject position from before consumption is pushed to the stack.
type TestAny struct {
	N   byte
	Lbl Label
	jump
}

// End immediately completes the pattern as a match.
type End struct {
	basic
}

// Nop does nothing.
type Nop struct {
	basic
}

const (
	CaptureBegin byte = iota
	CaptureEnd
	CaptureFull
)

type Capture struct {
	Attr  byte
	Extra byte
	basic
}

// Choice2 is a deprecated instruction.
type Choice2 struct {
	Lbl  Label
	Back byte
	jump
}

type basic struct{}

func (b basic) insn() {}

type jump struct {
	basic
}

func (j jump) jumpt() {}

func (i Label) String() string {
	return fmt.Sprintf("L%v", i.Id)
}

func (i Char) String() string {
	return fmt.Sprintf("Char %v", string(i.Byte))
}

func (i Jump) String() string {
	return fmt.Sprintf("Jump %v", i.Lbl)
}

func (i Choice) String() string {
	return fmt.Sprintf("Choice %v", i.Lbl)
}

func (i Call) String() string {
	return fmt.Sprintf("Call %v", i.Lbl)
}

func (i Commit) String() string {
	return fmt.Sprintf("Commit %v", i.Lbl)
}

func (i Return) String() string {
	return "Return"
}

func (i Fail) String() string {
	return "Fail"
}

func (i Set) String() string {
	return fmt.Sprintf("Set %v", i.Chars)
}

func (i Any) String() string {
	return fmt.Sprintf("Any %v", i.N)
}

func (i PartialCommit) String() string {
	return fmt.Sprintf("PartialCommit %v", i.Lbl)
}

func (i Span) String() string {
	return fmt.Sprintf("Span %v", i.Chars)
}

func (i BackCommit) String() string {
	return fmt.Sprintf("BackCommit %v", i.Lbl)
}

func (i FailTwice) String() string {
	return "FailTwice"
}

func (i TestChar) String() string {
	return fmt.Sprintf("TestChar %v %v", string(i.Byte), i.Lbl)
}

func (i TestSet) String() string {
	return fmt.Sprintf("TestSet %v %v", i.Chars, i.Lbl)
}

func (i TestAny) String() string {
	return fmt.Sprintf("TestAny %v %v", i.N, i.Lbl)
}

func (i Choice2) String() string {
	return fmt.Sprintf("Choice2 %v %v", i.Back, i.Lbl)
}

func (i End) String() string {
	return "End"
}

func (i Nop) String() string {
	return "Nop"
}

func (i Capture) String() string {
	switch i.Attr {
	case CaptureBegin:
		return fmt.Sprintf("Capture begin %v", i.Extra)
	case CaptureEnd:
		return fmt.Sprintf("Capture end")
	case CaptureFull:
		return fmt.Sprintf("Capture full %v", i.Extra)
	}
	return "Capture unknown"
}
