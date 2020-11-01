package vm

import (
	"bytes"
	"encoding/gob"
	"log"
	"unsafe"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
	"github.com/zyedidia/gpeg/pattern"
)

// VMCode is the representation of VM bytecode.
type VMCode struct {
	data code
}

type code struct {
	// list of charsets
	Sets []charset.Set

	// the encoded instructions
	Insns []byte
}

func (c *VMCode) Size() int {
	return int(unsafe.Sizeof(charset.Set{}))*len(c.data.Sets) + len(c.data.Insns)
}

func (c *VMCode) Bytes() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(c.data)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func LoadCode(b []byte) VMCode {
	var c code
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(&c)
	if err != nil {
		log.Fatal(err)
	}
	return VMCode{
		data: c,
	}
}

// Encode transforms a Pattern into VM bytecode.
func Encode(insns pattern.Pattern) VMCode {
	insns.Optimize()

	code := VMCode{
		data: code{
			Sets:  make([]charset.Set, 0),
			Insns: make([]byte, 0),
		},
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
			args = encodeU8(addSet(&code, t.Chars))
		case isa.Any:
			op = opAny
			args = []byte{t.N}
		case isa.PartialCommit:
			op = opPartialCommit
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.Span:
			op = opSpan
			args = encodeU8(addSet(&code, t.Chars))
		case isa.BackCommit:
			op = opBackCommit
			args = encodeLabel(labels[t.Lbl], sz)
		case isa.FailTwice:
			op = opFailTwice
		case isa.TestChar:
			op = opTestChar
			args = append([]byte{t.Byte}, encodeLabel(labels[t.Lbl], sz)...)
		case isa.TestCharNoChoice:
			op = opTestCharNoChoice
			args = append([]byte{t.Byte}, encodeLabel(labels[t.Lbl], sz)...)
		case isa.TestSet:
			op = opTestSet
			args = append(encodeU8(addSet(&code, t.Chars)), encodeLabel(labels[t.Lbl], sz)...)
		case isa.TestAny:
			op = opTestAny
			args = append([]byte{t.N}, encodeLabel(labels[t.Lbl], sz)...)
		case isa.CaptureBegin:
			op = opCaptureBegin
			args = encodeI16(int(t.Id))
		case isa.CaptureEnd:
			op = opCaptureEnd
		case isa.CaptureLate:
			op = opCaptureLate
			args = append([]byte{t.Back}, encodeI16(int(t.Id))...)
		case isa.CaptureFull:
			op = opCaptureFull
			args = append([]byte{t.Back}, encodeI16(int(t.Id))...)
		case isa.MemoOpen:
			op = opMemoOpen
			args = append(encodeLabel(labels[t.Lbl], sz), encodeI16(t.Id)...)
		case isa.MemoClose:
			op = opMemoClose
		case isa.End:
			op = opEnd
		default:
			continue
		}

		sz += size(insn)
		code.data.Insns = append(code.data.Insns, op)

		// need padding to align the args if they are divisible by 16 bits
		if len(args)%2 == 0 {
			code.data.Insns = append(code.data.Insns, 0)
		}

		code.data.Insns = append(code.data.Insns, args...)
	}
	code.data.Insns = append(code.data.Insns, opEnd, 0)

	return code
}

func encodeU8(x uint) []byte {
	return []byte{uint8(x)}
}

func encodeI8(x int) []byte {
	bytes := *(*[1]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeI16(x int) []byte {
	bytes := *(*[2]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeLabel(x int, cur int) []byte {
	return encodeI16(x - cur)
}

// Adds the set to the code's list of charsets, and returns the index it was
// added at. If there are duplicate charsets, this may not actually insert
// the new charset.
func addSet(code *VMCode, set charset.Set) uint {
	for i, s := range code.data.Sets {
		if set == s {
			return uint(i)
		}
	}

	code.data.Sets = append(code.data.Sets, set)
	return uint(len(code.data.Sets) - 1)
}
