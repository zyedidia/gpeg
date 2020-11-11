package vm

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"unsafe"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
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

func (c *VMCode) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(c.data)
	return buf.Bytes(), err
}

func FromBytes(b []byte) (VMCode, error) {
	var c code
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(&c)
	return VMCode{
		data: c,
	}, err
}

func (c *VMCode) ToJson() ([]byte, error) {
	return json.Marshal(c.data)
}

func FromJson(b []byte) (VMCode, error) {
	var c code
	err := json.Unmarshal(b, &c)
	return VMCode{
		data: c,
	}, err
}

// Encode transforms a Pattern into VM bytecode.
func Encode(insns isa.Program) VMCode {
	code := VMCode{
		data: code{
			Sets:  make([]charset.Set, 0),
			Insns: make([]byte, 0),
		},
	}

	var bcount uint
	labels := make(map[isa.Label]uint)
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
			args = encodeLabel(labels[t.Lbl])
		case isa.Choice:
			op = opChoice
			args = encodeLabel(labels[t.Lbl])
		case isa.Call:
			op = opCall
			args = encodeLabel(labels[t.Lbl])
		case isa.Commit:
			op = opCommit
			args = encodeLabel(labels[t.Lbl])
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
			args = encodeLabel(labels[t.Lbl])
		case isa.Span:
			op = opSpan
			args = encodeU8(addSet(&code, t.Chars))
		case isa.BackCommit:
			op = opBackCommit
			args = encodeLabel(labels[t.Lbl])
		case isa.FailTwice:
			op = opFailTwice
		case isa.TestChar:
			op = opTestChar
			args = append([]byte{t.Byte}, encodeLabel(labels[t.Lbl])...)
		case isa.TestCharNoChoice:
			op = opTestCharNoChoice
			args = append([]byte{t.Byte}, encodeLabel(labels[t.Lbl])...)
		case isa.TestSet:
			op = opTestSet
			args = append(encodeU8(addSet(&code, t.Chars)), encodeLabel(labels[t.Lbl])...)
		case isa.TestAny:
			op = opTestAny
			args = append([]byte{t.N}, encodeLabel(labels[t.Lbl])...)
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
			args = append(encodeI16(int(t.Id)), encodeLabel(labels[t.Lbl])...)
		case isa.MemoClose:
			op = opMemoClose
		case isa.End:
			op = opEnd
		default:
			continue
		}

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
	if x >= 256 {
		panic("U8 out of bounds")
	}

	return []byte{uint8(x)}
}

func encodeI8(x int) []byte {
	if x < -128 || x >= 128 {
		panic("I8 out of bounds")
	}

	bytes := *(*[1]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeU16(x uint) []byte {
	if x >= (1 << 16) {
		panic("U16 out of bounds")
	}

	bytes := *(*[2]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeI16(x int) []byte {
	if x < -(1<<15) || x >= (1<<15) {
		panic("I16 out of bounds")
	}

	bytes := *(*[2]byte)(unsafe.Pointer(&x))
	return bytes[:]
}

func encodeU24(x uint) []byte {
	if x >= (1 << 24) {
		panic("I24 out of bounds")
	}

	i1 := (x >> 16) & 0xff
	i2 := int16(x)

	bytes1 := *(*[1]byte)(unsafe.Pointer(&i1))
	bytes2 := *(*[2]byte)(unsafe.Pointer(&i2))
	bytes := append(bytes1[:], bytes2[:]...)
	return bytes
}

func encodeLabel(x uint) []byte {
	return encodeU24(x)
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
