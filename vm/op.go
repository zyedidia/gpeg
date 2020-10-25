package vm

import "github.com/zyedidia/gpeg/isa"

const (
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
)

func size(insn isa.Insn) int {
	var sz int
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
	case isa.Set, isa.TestSet, isa.Span, isa.MemoOpen:
		sz += 2
	}

	return sz
}

// size in bytes of each instruction's encoding (unused)
var sizes = map[byte]int{
	opChar:          2,
	opJump:          4,
	opChoice:        4,
	opCall:          4,
	opCommit:        4,
	opReturn:        2,
	opFail:          2,
	opSet:           4,
	opAny:           2,
	opPartialCommit: 4,
	opSpan:          4,
	opBackCommit:    4,
	opFailTwice:     2,
	opTestChar:      4,
	opTestSet:       6,
	opTestAny:       4,
	opEnd:           2,
	opNop:           0,
	opCaptureBegin:  2,
	opCaptureLate:   2,
	opCaptureEnd:    2,
	opCaptureFull:   2,
	opMemoOpen:      6,
	opMemoClose:     2,
}
