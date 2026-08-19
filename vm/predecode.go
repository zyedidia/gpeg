package vm

import "github.com/zyedidia/gpeg/charset"

// insn is one pre-decoded instruction. The byte-encoded form stays the
// wire/serialization format; before execution the program is decoded once
// into this fixed-size form so the interpreter loop does no operand
// decoding, no jump-offset arithmetic, and no charset table indexing per
// executed instruction.
type insn struct {
	op byte
	// a holds the primary operand: the jump target (as an instruction
	// index) for control-flow ops, otherwise the instruction's value
	// (char, count, id, error/checker index).
	a int32
	// b holds the secondary operand where one exists (char/count for the
	// test instructions, back amount for captures, flag/id pairs).
	b int32
	// set points into the code's charset table for set instructions.
	set *charset.Set
}

// opSizes is the encoded byte size of each opcode (see the sz* constants
// and the padding rule in Encode).
var opSizes = [...]int{
	opChar:             szChar,
	opJump:             szJump,
	opChoice:           szChoice,
	opCall:             szCall,
	opCommit:           szCommit,
	opReturn:           szReturn,
	opFail:             szFail,
	opSet:              szSet,
	opAny:              szAny,
	opPartialCommit:    szPartialCommit,
	opSpan:             szSpan,
	opBackCommit:       szBackCommit,
	opFailTwice:        szFailTwice,
	opEmpty:            szEmpty,
	opTestChar:         szTestChar,
	opTestCharNoChoice: szTestCharNoChoice,
	opTestSet:          szTestSet,
	opTestSetNoChoice:  szTestSetNoChoice,
	opTestAny:          szTestAny,
	opEnd:              szEnd,
	opNop:              2, // never emitted by Encode
	opCaptureBegin:     szCaptureBegin,
	opCaptureLate:      szCaptureLate,
	opCaptureEnd:       szCaptureEnd,
	opCaptureFull:      szCaptureFull,
	opCheckBegin:       szCheckBegin,
	opCheckEnd:         szCheckEnd,
	opMemoOpen:         szMemoOpen,
	opMemoClose:        szMemoClose,
	opMemoTreeOpen:     szMemoTreeOpen,
	opMemoTreeInsert:   szMemoTreeInsert,
	opMemoTree:         szMemoTree,
	opMemoTreeClose:    szMemoTreeClose,
	opError:            szError,
}

// predecode builds the pre-decoded instruction list from the encoded
// bytes. It is called by the Code constructors; exec also calls it lazily
// for package-internal hand-built Code values.
func (c *Code) predecode() {
	insns := c.data.Insns

	// First pass: decode operands, recording the instruction index each
	// byte offset maps to (for jump-target translation).
	idx := make([]int32, len(insns)+1)
	prog := make([]insn, 0, len(insns)/2)
	for ip := 0; ip < len(insns); {
		op := insns[ip]
		idx[ip] = int32(len(prog))
		in := insn{op: op}
		switch op {
		case opChar, opAny, opEmpty, opEnd:
			in.a = int32(insns[ip+1])
		case opJump, opChoice, opCall, opCommit, opPartialCommit,
			opBackCommit, opError, opCheckEnd:
			in.a = int32(decodeU24(insns[ip+1:]))
		case opSet, opSpan:
			in.set = &c.data.Sets[insns[ip+1]]
		case opTestChar, opTestCharNoChoice, opTestAny:
			in.b = int32(insns[ip+2])
			in.a = int32(decodeU24(insns[ip+3:]))
		case opTestSet, opTestSetNoChoice:
			in.set = &c.data.Sets[insns[ip+2]]
			in.a = int32(decodeU24(insns[ip+3:]))
		case opCaptureBegin, opMemoTreeClose:
			in.a = int32(decodeI16(insns[ip+2:]))
		case opCaptureLate, opCaptureFull:
			in.b = int32(insns[ip+1])
			in.a = int32(decodeI16(insns[ip+2:]))
		case opCheckBegin:
			in.a = int32(decodeI16(insns[ip+2:]))
			in.b = int32(decodeI16(insns[ip+4:]))
		case opMemoOpen, opMemoTreeOpen:
			in.a = int32(decodeU24(insns[ip+1:]))
			in.b = int32(decodeI16(insns[ip+4:]))
		}
		prog = append(prog, in)
		ip += opSizes[op]
	}
	idx[len(insns)] = int32(len(prog))

	// Second pass: translate jump targets from byte offsets to
	// instruction indices.
	for i := range prog {
		switch prog[i].op {
		case opJump, opChoice, opCall, opCommit, opPartialCommit,
			opBackCommit, opTestChar, opTestCharNoChoice, opTestSet,
			opTestSetNoChoice, opTestAny, opMemoOpen, opMemoTreeOpen:
			prog[i].a = idx[prog[i].a]
		}
	}
	c.prog = prog
}
