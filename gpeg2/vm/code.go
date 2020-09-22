package vm

import (
	"log"
	"unsafe"

	"github.com/zyedidia/gpeg/isa"
	"github.com/zyedidia/gpeg/pattern"
)

type VMCode []byte

func Encode(insns pattern.Pattern) VMCode {
	insns.Optimize()

	var code []byte

	// for label resolution
	var bcount uint32
	labels := make(map[int]uint32)
	for _, insn := range insns {
		switch t := insn.(type) {
		case isa.Label:
			labels[t.Id] = bcount
			continue
		case isa.JumpType:
			// op is 1 btye, label is 4 bytes
			bcount += 5
		default:
			// op is 1 byte
			bcount += 1
		}

		// handle extra arg sizes
		switch insn.(type) {
		case isa.Char, isa.Any, isa.TestChar, isa.TestAny, isa.Choice2:
			// arg is 1 byte
			bcount += 1
		case isa.Set, isa.Span, isa.TestSet:
			// arg is 16 bytes
			bcount += 16
		}
	}

	for _, insn := range insns {
		var op byte
		var args []byte
		switch t := insn.(type) {
		case isa.Label:
			continue
		case isa.Char:
			op = opChar
			args = []byte{t.Byte}
		case isa.Jump:
			op = opJump
			args = encodeU32(labels[t.Lbl.Id])
		case isa.Choice:
			op = opChoice
			args = encodeU32(labels[t.Lbl.Id])
		case isa.Call:
			op = opCall
			args = encodeU32(labels[t.Lbl.Id])
		case isa.Commit:
			op = opCommit
			args = encodeU32(labels[t.Lbl.Id])
		case isa.Return:
			op = opReturn
		case isa.Fail:
			op = opFail
		case isa.Set:
			op = opSet
			args = encodeSet(t.Chars)
		case isa.Any:
			op = opAny
			args = []byte{t.N}
		case isa.PartialCommit:
			op = opPartialCommit
			args = encodeU32(labels[t.Lbl.Id])
		case isa.Span:
			op = opSpan
			args = encodeSet(t.Chars)
		case isa.BackCommit:
			op = opBackCommit
			args = encodeU32(labels[t.Lbl.Id])
		case isa.FailTwice:
			op = opFailTwice
		case isa.TestChar:
			op = opTestChar
			args = append(encodeU32(labels[t.Lbl.Id]), t.Byte)
		case isa.TestSet:
			op = opTestSet
			args = append(encodeU32(labels[t.Lbl.Id]), encodeSet(t.Chars)...)
		case isa.TestAny:
			op = opTestAny
			args = append(encodeU32(labels[t.Lbl.Id]), t.N)
		case isa.Choice2:
			op = opChoice2
			args = append(encodeU32(labels[t.Lbl.Id]), t.Back)
		case isa.End:
			op = opEnd
		default:
			log.Println("Invalid instruction", t)
			continue
		}
		code = append(code, op)
		code = append(code, args...)
	}

	return code
}

func encodeSet(set isa.Charset) []byte {
	bytes := *(*[16]byte)(unsafe.Pointer(&set.Bits[0]))
	return bytes[:]
	// var b []byte
	// b = append(b, encodeU32(uint32(set.Bits[0]&0xffffffff))...)
	// b = append(b, encodeU32(uint32((set.Bits[0]>>32)&0xffffffff))...)
	// b = append(b, encodeU32(uint32(set.Bits[1]&0xffffffff))...)
	// b = append(b, encodeU32(uint32((set.Bits[1]>>32)&0xffffffff))...)
	// return b
}

func encodeU32(l uint32) []byte {
	bytes := *(*[4]byte)(unsafe.Pointer(&l))
	return bytes[:]
	// b1 := byte(l & 0xff)
	// b2 := byte((l >> 8) & 0xff)
	// b3 := byte((l >> 16) & 0xff)
	// b4 := byte((l >> 24) & 0xff)
	// return []byte{b1, b2, b3, b4}
}
