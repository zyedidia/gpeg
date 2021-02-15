package vm

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"

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
	// list of error messages
	Errors []string

	// the encoded instructions
	Insns []byte
}

// Size returns the size of the encoded instructions.
func (c *VMCode) Size() int {
	return len(c.data.Insns)
}

// ToBytes serializes and compresses this VMCode.
func (c *VMCode) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	fz := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(fz)
	err := enc.Encode(c.data)
	fz.Close()
	return buf.Bytes(), err
}

// FromBytes loads a VMCode from a compressed and serialized object.
func FromBytes(b []byte) (VMCode, error) {
	var c code
	fz, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return VMCode{}, err
	}
	dec := gob.NewDecoder(fz)
	err = dec.Decode(&c)
	fz.Close()
	return VMCode{
		data: c,
	}, err
}

// ToJson returns this VMCode serialized to JSON form.
func (c *VMCode) ToJson() ([]byte, error) {
	return json.Marshal(c.data)
}

// FromJson returns a VMCode loaded from JSON form.
func FromJson(b []byte) (VMCode, error) {
	var c code
	err := json.Unmarshal(b, &c)
	return VMCode{
		data: c,
	}, err
}

// Encode transforms a program into VM bytecode.
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
		case isa.TestSetNoChoice:
			op = opTestSetNoChoice
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
			args = append(encodeLabel(labels[t.Lbl]), encodeI16(int(t.Id))...)
		case isa.MemoClose:
			op = opMemoClose
		case isa.MemoTree:
			op = opMemoTree
		case isa.MemoTreeClose:
			op = opMemoTreeClose
		case isa.Error:
			op = opError
			args = encodeU24(addError(&code, t.Message))
		case isa.End:
			op = opEnd
			args = encodeBool(t.Fail)
		default:
			panic(fmt.Sprintf("invalid instruction during encoding: %v", t))
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

	return []byte{byte(x)}
}

func encodeU16(x uint) []byte {
	if x >= (1 << 16) {
		panic("U16 out of bounds")
	}

	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b[0:], uint16(x))
	return b
}

func encodeI16(x int) []byte {
	if x < -(1<<15) || x >= (1<<15) {
		panic("I16 out of bounds")
	}

	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b[0:], uint16(x))
	return b
}

func encodeU24(x uint) []byte {
	if x >= (1 << 24) {
		panic("I24 out of bounds")
	}

	b := make([]byte, 4)
	i1 := uint16((x >> 16) & 0xff)
	i2 := uint16(x)

	binary.LittleEndian.PutUint16(b[0:], i1)
	binary.LittleEndian.PutUint16(b[2:], i2)
	return b[1:4]
}

func encodeLabel(x uint) []byte {
	return encodeU24(x)
}

func encodeBool(b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
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

func addError(code *VMCode, msg string) uint {
	for i, s := range code.data.Errors {
		if msg == s {
			return uint(i)
		}
	}

	code.data.Errors = append(code.data.Errors, msg)
	return uint(len(code.data.Errors) - 1)
}
