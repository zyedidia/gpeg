package vm

import (
	"unsafe"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
	"github.com/zyedidia/gpeg/pattern"
)

// VMCode is the representation of VM bytecode.
type VMCode struct {
	// list of charsets
	sets []charset.Set

	// the encoded instructions
	insns []byte
}

// Encode transforms a Pattern into VM bytecode.
func Encode(insns pattern.Pattern) VMCode {
	insns.Optimize()

	code := VMCode{
		sets:  make([]charset.Set, 0),
		insns: make([]byte, 0),
	}

	var bcount int
	labels := make(map[isa.Label]int)
	for _, insn := range insns {
		switch t := insn.(type) {
		case isa.Nop:
			continue
		case isa.Label:
			labels[t] = bcount
			continue
		default:
			bcount += size(insn)
		}
	}

	var sz int
	for _, insn := range insns {
		var op byte
		var args []byte

		switch t := insn.(type) {
		case isa.Label, isa.Nop:
			continue
		case isa.Char:
			op = opChar
			args = []byte{t.Byte}
		case isa.Jump:
			op = opJump
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Choice:
			op = opChoice
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Call:
			op = opCall
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Commit:
			op = opCommit
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Return:
			op = opReturn
		case isa.Fail:
			op = opFail
		case isa.Set:
			op = opSet
			args = encodeU16(addSet(&code, t.Chars))
		case isa.Any:
			op = opAny
			args = []byte{t.N}
		case isa.PartialCommit:
			op = opPartialCommit
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Span:
			op = opSpan
			args = encodeU16(addSet(&code, t.Chars))
		case isa.BackCommit:
			op = opBackCommit
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.FailTwice:
			op = opFailTwice
		case isa.TestChar:
			op = opTestChar
			args = append([]byte{t.Byte}, encodeLabel(labels[t.Lbl], sz)...)
		case isa.TestSet:
			op = opTestSet
			args = append(encodeLabel(labels[t.Lbl], sz), encodeU16(addSet(&code, t.Chars))...)
		case isa.TestAny:
			op = opTestAny
			args = append([]byte{t.N}, encodeLabel(labels[t.Lbl], sz)...)
		case isa.CaptureBegin:
			op = opCaptureBegin
		case isa.CaptureEnd:
			op = opCaptureEnd
		case isa.CaptureLate:
			op = opCaptureLate
			args = []byte{t.N}
		case isa.CaptureFull:
			op = opCaptureFull
			args = []byte{t.N}
		case isa.MemoOpen:
			op = opMemoOpen
			args = append(encodeLabel(labels[t.Lbl], sz), encodeU16(t.Id)...)
		case isa.MemoClose:
			op = opMemoClose
		case isa.End:
			op = opEnd
		default:
			continue
		}

		sz += size(insn)
		code.insns = append(code.insns, op)

		// need padding to align the args if they are divisible by 16 bits
		if len(args)%2 == 0 {
			code.insns = append(code.insns, 0)
		}

		code.insns = append(code.insns, args...)
	}
	code.insns = append(code.insns, opEnd, 0)

	return code
}

func encodeU16(x int) []byte {
	bytes := *(*[2]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeLabel(x int, cur int) []byte {
	return encodeU16(x - cur)
}

// Adds the set to the code's list of charsets, and returns the index it was
// added at. If there are duplicate charsets, this may not actually insert
// the new charset.
func addSet(code *VMCode, set charset.Set) int {
	for i, s := range code.sets {
		if set == s {
			return i
		}
	}

	code.sets = append(code.sets, set)
	return len(code.sets) - 1
}
