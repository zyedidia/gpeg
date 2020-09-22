package vm

import (
	"log"

	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/isa"
)

const ipFail = -1

type VM struct {
	ip    int
	st    *stack
	start input.Pos
	input *input.BufferedReader
}

func NewVM(r input.Reader, start input.Pos) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		start: start,
		input: input.NewBufferedReader(r, start),
	}
}

func (vm *VM) Exec(code VMCode) int {
loop:
	for vm.ip < len(code) {
		if vm.ip == ipFail {
			ent, ok := vm.st.pop()
			if !ok {
				// match failed
				return -1
			}
			if !ent.isRet() {
				vm.ip = ent.btrack.ip
				vm.input.SeekTo(ent.btrack.off)
			}
			// try again with new ip/stack
			continue
		}

		op := code[vm.ip]
		switch op {
		case opChar:
			b := decodeByte(code[vm.ip+1:])
			in, eof := vm.input.Peek()
			if eof == nil && b == in {
				vm.input.Advance(1)
				vm.ip += 2
			} else {
				vm.ip = ipFail
			}
		case opJump:
			lbl := decodeU32(code[vm.ip+1:])
			vm.ip = int(lbl)
		case opChoice:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.push(vm.st.backtrack(int(lbl), vm.input.Offset()))
			vm.ip += 5
		case opCall:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.push(vm.st.retaddr(vm.ip + 5))
			vm.ip = int(lbl)
		case opCommit:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.pop()
			vm.ip = int(lbl)
		case opReturn:
			ent, ok := vm.st.pop()
			if ok && ent.isRet() {
				vm.ip = ent.retaddr
			} else {
				panic("Return failed")
			}
		case opFail:
			vm.ip = ipFail
		case opSet:
			set := decodeSet(code[vm.ip+1:])
			in, eof := vm.input.Peek()
			if eof == nil && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += 17
			} else {
				vm.ip = ipFail
			}
		case opAny:
			n := decodeByte(code[vm.ip+1:])
			err := vm.input.Advance(int(n))
			if err != nil {
				vm.ip = ipFail
			} else {
				vm.ip += 2
			}
		case opPartialCommit:
			lbl := decodeU32(code[vm.ip+1:])
			ent := vm.st.peek()
			if ent != nil && !ent.isRet() {
				ent.btrack.off = vm.input.Offset()
				vm.ip = int(lbl)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			set := decodeSet(code[vm.ip+1:])
			in, eof := vm.input.Peek()
			for eof == nil && set.Has(in) {
				vm.input.Advance(1)
				in, eof = vm.input.Peek()
			}
			vm.ip += 17
		case opBackCommit:
			lbl := decodeU32(code[vm.ip+1:])
			ent, ok := vm.st.pop()
			if ok && !ent.isRet() {
				vm.input.SeekTo(ent.btrack.off)
				vm.ip = int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			vm.st.pop()
			vm.ip = ipFail
		case opTestChar:
			lbl := decodeU32(code[vm.ip+1:])
			b := decodeByte(code[vm.ip+1+4:])
			in, eof := vm.input.Peek()
			if eof == nil && in == b {
				vm.input.Advance(1)
				vm.ip += 6
			} else {
				vm.ip = int(lbl)
			}
		case opTestSet:
			lbl := decodeU32(code[vm.ip+1:])
			set := decodeSet(code[vm.ip+1+4:])
			in, eof := vm.input.Peek()
			if eof == nil && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += 21
			} else {
				vm.ip = int(lbl)
			}
		case opTestAny:
			lbl := decodeU32(code[vm.ip+1:])
			n := decodeByte(code[vm.ip+1+4:])
			err := vm.input.Advance(int(n))
			if err != nil {
				vm.ip = int(lbl)
			} else {
				vm.ip += 6
			}
		case opEnd:
			// ends the machine with a success
			break loop
		case opChoice2:
			// lbl := decodeU32(code[vm.ip+1:])
			// back := decodeByte(code[vm.ip+1+4:])
		default:
			log.Fatal("Invalid opcode")
		}
	}

	return vm.input.Offset().Distance(vm.start)
}

func decodeByte(b []byte) byte {
	return b[0]
}

func decodeU32(b []byte) uint32 {
	return uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func decodeSet(b []byte) isa.Charset {
	first := uint64(decodeU32(b)) | (uint64(decodeU32(b[4:])) << 32)
	second := uint64(decodeU32(b[8:])) | (uint64(decodeU32(b[12:])) << 32)

	return isa.Charset{
		Bits: [2]uint64{first, second},
	}
}
