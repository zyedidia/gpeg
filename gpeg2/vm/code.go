package vm

import "github.com/zyedidia/gpeg/isa"

type VMCode struct {
	code []byte
}

func (v VMCode) Serialize() []byte {
	return v.code
}

func Deserialize(b []byte) VMCode {
	return VMCode{
		code: b,
	}
}

func Encode(insns []isa.Insn) VMCode {
	var code []byte

	// resolve labels
	icount := 0
	labels := make(map[int]int)
	for _, insn := range insns {
		switch t := insn.(type) {
		case isa.Label:
			labels[t.Id] = icount + 1
		default:
			icount++
		}
	}

	// generate vm code
	for _, insn := range insns {
		code = append(code, encodeInsn(insn, labels)...)
	}

	return VMCode{
		code: code,
	}
}

func encodeInsn(insn isa.Insn, labels map[int]int) []byte {
	switch t := insn.(type) {
	case isa.Char:
	case isa.Jump:
	case isa.Choice:
	case isa.Call:
	case isa.Commit:
	case isa.Return:
	case isa.Fail:
	case isa.Set:
	case isa.Any:
	case isa.PartialCommit:
	case isa.Span:
	case isa.BackCommit:
	case isa.FailTwice:
	case isa.TestChar:
	case isa.TestSet:
	case isa.TestAny:
	case isa.Choice2:
	}
	return []byte{}
}

func setToBytes(set isa.Charset) []byte {

}
