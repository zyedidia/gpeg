package vm

import (
	"github.com/zyedidia/gpeg/isa"
)

const (
	// base instruction set
	opChar byte = iota
	opJump
	opChoice
	opCall
	opCommit
	opReturn
	opFail
	opSet
	opAny
	opPartialCommit
	opSpan
	opBackCommit
	opFailTwice
	opTestChar
	opTestCharNoChoice
	opTestSet
	opTestSetNoChoice
	opTestAny
	opEnd
	opNop
	opCaptureBegin
	opCaptureLate
	opCaptureEnd
	opCaptureFull
	opMemoOpen
	opMemoClose
)

// instruction sizes
const (
	// base instruction set
	szChar         = 2
	szReturn       = 2
	szFail         = 2
	szSet          = 2
	szAny          = 2
	szSpan         = 2
	szFailTwice    = 2
	szEnd          = 2
	szNop          = 0
	szCaptureBegin = 4
	szCaptureLate  = 4
	szCaptureEnd   = 2
	szCaptureFull  = 4
	szMemoClose    = 2

	// jumps
	szJump             = 4
	szChoice           = 4
	szCall             = 4
	szCommit           = 4
	szPartialCommit    = 4
	szBackCommit       = 4
	szTestChar         = 6
	szTestCharNoChoice = 6
	szTestSet          = 6
	szTestSetNoChoice  = 6
	szTestAny          = 6
	szMemoOpen         = 6
)

// returns the size in bytes of the encoded version of this instruction
func size(insn isa.Insn) uint {
	var sz uint
	switch insn.(type) {
	case isa.Label, isa.Nop:
		return 0
	case isa.JumpType:
		sz += 4
	default:
		sz += 2
	}

	// handle instructions with extra args
	switch insn.(type) {
	case isa.MemoOpen, isa.CaptureBegin, isa.CaptureLate, isa.CaptureFull,
		isa.TestChar, isa.TestCharNoChoice, isa.TestSet, isa.TestSetNoChoice, isa.TestAny:
		sz += 2
	}

	return sz
}
