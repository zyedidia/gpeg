package vm

import (
	"fmt"

	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/isa"
	"github.com/zyedidia/gpeg/memo"
)

const memoSize = 100

// A VM represents the state for the virtual machine. It stores a reference
// to the input reader, an instruction pointer, a stack of backtrack entries
// and return address entries, and the initial subject position (intermediate
// subject positions are stored on the stack as backtrack entries).
type VM struct {
	ip   int
	st   *stack
	capt []capt
	memo *memo.Table

	start input.Pos
	input *input.BufferedReader
}

// NewVM returns a new virtual machine which will read from the given
// input.Reader starting at the start position.
func NewVM(r input.Reader, start input.Pos) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		start: start,
		input: input.NewBufferedReader(r, start),
		capt:  []capt{},
		memo:  memo.NewTable(memoSize),
	}
}

// Reset resets the VM state to initial values and the given start position in
// the subject.
func (vm *VM) Reset(start input.Pos) {
	vm.ip = 0
	vm.start = start
	vm.memo = memo.NewTable(memoSize)
	vm.input.SeekTo(vm.start)
	vm.st.reset()
}

// Exec executes the given VM bytecode using the current VM state and returns
// whether the match was successful, the offset when matching stopped, and a
// capture object.
func (vm *VM) Exec(code VMCode) (bool, input.Pos, []capt) {
loop:
	for {
		op := code[vm.ip]
		switch op {
		case opChar:
			b := decodeByte(code[vm.ip+1:])
			in, ok := vm.input.Peek()
			if ok && b == in {
				vm.input.Advance(1)
				vm.ip += 2
			} else {
				goto fail
			}
		case opJump:
			lbl := decodeU32(code[vm.ip+1:])
			vm.ip = int(lbl)
		case opChoice:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.push(stackBacktrack{int(lbl), vm.input.Offset(), vm.capt})
			vm.ip += 5
		case opCall:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.push(stackRet(vm.ip + 5))
			vm.ip = int(lbl)
		case opCommit:
			lbl := decodeU32(code[vm.ip+1:])
			vm.st.pop()
			vm.ip = int(lbl)
		case opReturn:
			ent := vm.st.pop()
			if addr, isret := ent.(stackRet); isret {
				vm.ip = int(addr)
			} else {
				panic("Return failed")
			}
		case opFail:
			goto fail
		case opSet:
			set := decodeSet(code[vm.ip+1:])
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += 17
			} else {
				goto fail
			}
		case opAny:
			n := decodeByte(code[vm.ip+1:])
			ok := vm.input.Advance(int(n))
			if ok {
				vm.ip += 2
			} else {
				goto fail
			}
		case opPartialCommit:
			lbl := decodeU32(code[vm.ip+1:])
			ent := vm.st.peek()
			// TODO: does this work?
			if btrack, ok := (*ent).(stackBacktrack); ok {
				btrack.off = vm.input.Offset()
				btrack.capt = vm.capt
				*ent = btrack
				vm.ip = int(lbl)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			set := decodeSet(code[vm.ip+1:])
			in, ok := vm.input.Peek()
			for ok && set.Has(in) {
				vm.input.Advance(1)
				in, ok = vm.input.Peek()
			}
			vm.ip += 17
		case opBackCommit:
			lbl := decodeU32(code[vm.ip+1:])
			ent := vm.st.pop()
			if btrack, ok := ent.(stackBacktrack); ok {
				vm.input.SeekTo(btrack.off)
				vm.capt = btrack.capt
				vm.ip = int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			vm.st.pop()
			goto fail
		case opTestChar:
			lbl := decodeU32(code[vm.ip+1:])
			b := decodeByte(code[vm.ip+1+4:])
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.st.push(stackBacktrack{int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += 6
			} else {
				vm.ip = int(lbl)
			}
		case opTestSet:
			lbl := decodeU32(code[vm.ip+1:])
			set := decodeSet(code[vm.ip+1+4:])
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.st.push(stackBacktrack{int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += 21
			} else {
				vm.ip = int(lbl)
			}
		case opTestAny:
			lbl := decodeU32(code[vm.ip+1:])
			n := decodeByte(code[vm.ip+1+4:])
			ent := stackBacktrack{int(lbl), vm.input.Offset(), vm.capt}
			ok := vm.input.Advance(int(n))
			if ok {
				vm.st.push(ent)
				vm.ip += 6
			} else {
				vm.ip = int(lbl)
			}
		case opCapture:
			c := capt{
				ip:  vm.ip,
				off: vm.input.Offset(),
			}
			vm.capt = append(vm.capt, c)
			vm.ip += 3
		case opEnd:
			// ends the machine with a success
			break loop
		case opChoice2:
			lbl := decodeU32(code[vm.ip+1:])
			back := decodeByte(code[vm.ip+1+4:])
			vm.st.push(stackBacktrack{int(lbl), vm.input.Offset() - input.Pos(back), vm.capt})
			vm.ip += 6
		case opMemoOpen:
			lbl := decodeU32(code[vm.ip+1:])
			id := decodeU16(code[vm.ip+1+4:])

			ment, ok := vm.memo.Get(memo.Key{
				Id:  id,
				Pos: vm.input.Offset(),
			})
			if ok {
				if ment.MatchLength() == -1 {
					goto fail
				}
				vm.input.Advance(ment.MatchLength())
				vm.ip = int(lbl)
			} else {
				vm.st.push(stackMemo{
					id:  id,
					pos: vm.input.Offset(),
				})
				vm.ip += 7
			}
		case opMemoClose:
			ent := vm.st.pop()
			if ment, ok := ent.(stackMemo); ok {
				vm.memo.Put(memo.Key{
					Id:  ment.id,
					Pos: ment.pos,
				}, memo.NewEntry(int(vm.input.Offset())-int(ment.pos), 0))
				vm.ip += 1
			} else {
				panic("MemoClose found no partial memo entry!")
			}
		case opNop:
			vm.ip += 1
		default:
			panic("Invalid opcode")
		}
	}

	fmt.Println(vm.memo)

	return true, vm.input.Offset(), vm.capt

fail:
	ent := vm.st.pop()
	if ent == nil {
		// match failed
		return false, vm.input.Offset(), vm.capt
	}

	switch t := ent.(type) {
	case stackBacktrack:
		vm.ip = t.ip
		vm.input.SeekTo(t.off)
		vm.capt = t.capt
	case stackMemo:
		// Mark this position in the memoTable as a failed match
		vm.memo.Put(memo.Key{
			Id:  t.id,
			Pos: t.pos,
		}, memo.NewEntry(-1, -1))
		goto fail
	case stackRet:
		goto fail
	}

	goto loop
}

func decodeByte(b []byte) byte {
	return b[0]
}

func decodeU32(b []byte) uint32 {
	return uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func decodeU16(b []byte) uint16 {
	return uint16(b[0]) | (uint16(b[1]) << 8)
}

func decodeSet(b []byte) isa.Charset {
	first := uint64(decodeU32(b)) | (uint64(decodeU32(b[4:])) << 32)
	second := uint64(decodeU32(b[8:])) | (uint64(decodeU32(b[12:])) << 32)

	return isa.Charset{
		Bits: [2]uint64{first, second},
	}
}
