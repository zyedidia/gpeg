package isa

import (
	"fmt"
	"strconv"

	"github.com/zyedidia/gpeg/charset"
)

// Insn represents the interface for an instruction in the ISA
type Insn interface {
	insn()
}

// A Program is a sequence of instructions
type Program []Insn

// Size returns the number of instructions in a program ignoring labels and
// nops.
func (p Program) Size() int {
	var sz int
	for _, i := range p {
		switch i.(type) {
		case Label, Nop:
			continue
		default:
			sz++
		}
	}
	return sz
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
	Chars charset.Set
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
	Chars charset.Set
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

// TestCharNoChoice consumes the next byte if it matches Byte and jumps to Lbl
// otherwise. No backtrack entry is pushed to the stack.
type TestCharNoChoice struct {
	Byte byte
	Lbl  Label
	jump
}

// TestSet consumes the next byte if it is in the set Chars and jumps to
// Lbl otherwise. If the consumption is possible, a backtrack entry referring
// to Lbl and the subject position from before consumption is pushed to the
// stack.
type TestSet struct {
	Chars charset.Set
	Lbl   Label
	jump
}

// TestSetNoChoice is the same as TestSet but no backtrack entry is pushed to
// the stack.
type TestSetNoChoice struct {
	Chars charset.Set
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

// MemoOpen begins a memo entry at this position. It marks the pattern that is
// being memoized with a unique ID for that pattern, and stores a label to
// jump to if the pattern is found in the memoization table.
type MemoOpen struct {
	Lbl Label
	Id  int16
	jump
}

// MemoClose completes a memoization entry and adds the entry into the memo
// table if it meets certain conditions (size, or other heuristics).
type MemoClose struct {
	basic
}

// CaptureBegin begins capturing the given ID.
type CaptureBegin struct {
	Id int16
	basic
}

// CaptureLate begins capturing the given ID at the current subject position
// minus Back.
type CaptureLate struct {
	Back byte
	Id   int16
	basic
}

// CaptureEnd completes an active capture.
type CaptureEnd struct {
	Id int16
	basic
}

// CaptureFull begins a capture for the given ID at the current subject
// position minus Back, and immediately completes the capture. This is
// equivalent to CaptureLate Back ID; CaptureEnd.
type CaptureFull struct {
	Back byte
	Id   int16
	basic
}

type basic struct{}

func (b basic) insn() {}

type jump struct {
	basic
}

func (j jump) jumpt() {}

// String returns the string representation of this instruction.
func (i Label) String() string {
	return fmt.Sprintf("L%v", i.Id)
}

// String returns the string representation of this instruction.
func (i Char) String() string {
	return fmt.Sprintf("Char %v", strconv.QuoteRune(rune(i.Byte)))
}

// String returns the string representation of this instruction.
func (i Jump) String() string {
	return fmt.Sprintf("Jump %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i Choice) String() string {
	return fmt.Sprintf("Choice %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i Call) String() string {
	return fmt.Sprintf("Call %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i Commit) String() string {
	return fmt.Sprintf("Commit %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i Return) String() string {
	return "Return"
}

// String returns the string representation of this instruction.
func (i Fail) String() string {
	return "Fail"
}

// String returns the string representation of this instruction.
func (i Set) String() string {
	return fmt.Sprintf("Set %v", i.Chars)
}

// String returns the string representation of this instruction.
func (i Any) String() string {
	return fmt.Sprintf("Any %v", i.N)
}

// String returns the string representation of this instruction.
func (i PartialCommit) String() string {
	return fmt.Sprintf("PartialCommit %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i Span) String() string {
	return fmt.Sprintf("Span %v", i.Chars)
}

// String returns the string representation of this instruction.
func (i BackCommit) String() string {
	return fmt.Sprintf("BackCommit %v", i.Lbl)
}

// String returns the string representation of this instruction.
func (i FailTwice) String() string {
	return "FailTwice"
}

// String returns the string representation of this instruction.
func (i TestChar) String() string {
	return fmt.Sprintf("TestChar %v %v", strconv.QuoteRune(rune(i.Byte)), i.Lbl)
}

// String returns the string representation of this instruction.
func (i TestCharNoChoice) String() string {
	return fmt.Sprintf("TestCharNoChoice %v %v", strconv.QuoteRune(rune(i.Byte)), i.Lbl)
}

// String returns the string representation of this instruction.
func (i TestSet) String() string {
	return fmt.Sprintf("TestSet %v %v", i.Chars, i.Lbl)
}

// String returns the string representation of this instruction.
func (i TestSetNoChoice) String() string {
	return fmt.Sprintf("TestSetNoChoice %v %v", i.Chars, i.Lbl)
}

// String returns the string representation of this instruction.
func (i TestAny) String() string {
	return fmt.Sprintf("TestAny %v %v", i.N, i.Lbl)
}

// String returns the string representation of this instruction.
func (i End) String() string {
	return "End"
}

// String returns the string representation of this instruction.
func (i Nop) String() string {
	return "Nop"
}

// String returns the string representation of this instruction.
func (i MemoOpen) String() string {
	return fmt.Sprintf("MemoOpen %v %v", i.Lbl, i.Id)
}

// String returns the string representation of this instruction.
func (i MemoClose) String() string {
	return "MemoClose"
}

// String returns the string representation of this instruction.
func (i CaptureBegin) String() string {
	return fmt.Sprintf("Capture begin %v", i.Id)
}

// String returns the string representation of this instruction.
func (i CaptureLate) String() string {
	return fmt.Sprintf("Capture late %v %v", i.Back, i.Id)
}

// String returns the string representation of this instruction.
func (i CaptureEnd) String() string {
	return fmt.Sprintf("Capture end")
}

// String returns the string representation of this instruction.
func (i CaptureFull) String() string {
	return fmt.Sprintf("Capture full %v %v", i.Back, i.Id)
}

// String returns the string representation of the program.
func (p Program) String() string {
	s := ""
	var last Insn
	for _, insn := range p {
		switch insn.(type) {
		case Nop:
			continue
		case Label:
			if _, ok := last.(Label); ok {
				s += "\rL...:"
			} else {
				s += fmt.Sprintf("%v:", insn)
			}
		default:
			s += fmt.Sprintf("\t%v\n", insn)
		}
		last = insn
	}
	s += "\n"
	return s
}
