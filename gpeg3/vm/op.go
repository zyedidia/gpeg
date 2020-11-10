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
	opTestAny
	opEnd
	opNop
	opCaptureBegin
	opCaptureLate
	opCaptureEnd
	opCaptureFull
	opMemoOpen
	opMemoClose

	// big jump variants
	opBigJump
	opBigChoice
	opBigCall
	opBigCommit
	opBigPartialCommit
	opBigBackCommit
	opBigTestChar
	opBigTestCharNoChoice
	opBigTestSet
	opBigTestAny
	opBigMemoOpen

	// small jump variants
	opSmallJump
	opSmallChoice
	opSmallCall
	opSmallCommit
	opSmallPartialCommit
	opSmallBackCommit
	opSmallTestSet
)

// returns the size in bytes of the encoded version of this instruction
func size(insn isa.Insn) int {
	var sz int
	switch insn.(type) {
	case isa.Label, isa.Nop:
		return 0
	case isa.JumpType:
		sz += 6
	default:
		sz += 2
	}

	// handle instructions with extra args
	switch insn.(type) {
	case isa.MemoOpen, isa.CaptureBegin, isa.CaptureLate, isa.CaptureFull:
		sz += 2
	}

	return sz
}

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

	// normal jumps
	szJump             = 4
	szChoice           = 4
	szCall             = 4
	szCommit           = 4
	szPartialCommit    = 4
	szBackCommit       = 4
	szTestChar         = 4
	szTestCharNoChoice = 4
	szTestSet          = 4
	szTestAny          = 4
	szMemoOpen         = 6

	// big jump variants
	szBigJump             = 6
	szBigChoice           = 6
	szBigCall             = 6
	szBigCommit           = 6
	szBigPartialCommit    = 6
	szBigBackCommit       = 6
	szBigTestChar         = 6
	szBigTestCharNoChoice = 6
	szBigTestSet          = 6
	szBigTestAny          = 6
	szBigMemoOpen         = 8

	// small jump variants
	szSmallJump          = 2
	szSmallChoice        = 2
	szSmallCall          = 2
	szSmallCommit        = 2
	szSmallPartialCommit = 2
	szSmallBackCommit    = 2
)
