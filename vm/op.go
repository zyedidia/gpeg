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
	opEmpty
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
	opCheckBegin
	opCheckEnd
	opMemoOpen
	opMemoClose
	opMemoTreeOpen
	opMemoTreeInsert
	opMemoTree
	opMemoTreeClose
	opError
)

// instruction sizes
const (
	// base instruction set
	szChar           = 2
	szReturn         = 2
	szFail           = 2
	szSet            = 2
	szAny            = 2
	szSpan           = 2
	szFailTwice      = 2
	szEnd            = 2
	szNop            = 0
	szEmpty          = 2
	szCaptureBegin   = 4
	szCaptureLate    = 4
	szCaptureEnd     = 2
	szCaptureFull    = 4
	szMemoClose      = 2
	szMemoTreeInsert = 2
	szMemoTree       = 2
	szMemoTreeClose  = 4
	szCheckBegin     = 6
	szCheckEnd       = 4
	szError          = 4

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
	szMemoTreeOpen     = 6
)

// returns the size in bytes of the encoded version of this instruction
func size(insn isa.Insn) uint {
	var sz uint
	switch insn.(type) {
	case isa.Label, isa.Nop:
		return 0
	case isa.JumpType, isa.CheckBegin:
		sz += 4
	default:
		sz += 2
	}

	// handle instructions with extra args
	switch insn.(type) {
	case isa.MemoOpen, isa.MemoTreeOpen, isa.MemoTreeClose, isa.CaptureBegin, isa.CaptureLate,
		isa.CaptureFull, isa.TestChar, isa.TestCharNoChoice, isa.TestSet,
		isa.TestSetNoChoice, isa.TestAny, isa.Error, isa.CheckBegin, isa.CheckEnd:
		sz += 2
	}

	return sz
}

var names = map[byte]string{
	opChar:             "Char",
	opJump:             "Jump",
	opChoice:           "Choice",
	opCall:             "Call",
	opCommit:           "Commit",
	opReturn:           "Return",
	opFail:             "Fail",
	opSet:              "Set",
	opAny:              "Any",
	opPartialCommit:    "PartialCommit",
	opSpan:             "Span",
	opBackCommit:       "BackCommit",
	opFailTwice:        "FailTwice",
	opTestChar:         "TestChar",
	opTestCharNoChoice: "TestCharNoChoice",
	opTestSet:          "TestSet",
	opTestSetNoChoice:  "TestSetNoChoice",
	opTestAny:          "TestAny",
	opEnd:              "End",
	opNop:              "Nop",
	opCaptureBegin:     "CaptureBegin",
	opCaptureLate:      "CaptureLate",
	opCaptureEnd:       "CaptureEnd",
	opCaptureFull:      "CaptureFull",
	opCheckBegin:       "CheckBegin",
	opCheckEnd:         "CheckEnd",
	opMemoOpen:         "MemoOpen",
	opMemoClose:        "MemoClose",
	opMemoTreeOpen:     "MemoTreeOpen",
	opMemoTreeInsert:   "MemoTreeInsert",
	opMemoTree:         "MemoTree",
	opMemoTreeClose:    "MemoTreeClose",
	opError:            "Error",
	opEmpty:            "Empty",
}

func opstr(op byte) string {
	return names[op]
}
