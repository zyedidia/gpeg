package vm

import (
	"unsafe"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
)

const memoCutoff = 128

type VM struct {
	ip   int
	st   *stack
	capt []capt
	code VMCode

	input *input.BufferedReader
}

func NewVM(r input.Reader, code VMCode) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		input: input.NewBufferedReader(r),
		capt:  []capt{},
		code:  code,
	}
}

func (vm *VM) SeekTo(p input.Pos) {
	vm.input.SeekTo(p)
}

func (vm *VM) Reset() {
	vm.ip = 0
	vm.input.ResetMaxExamined()
	vm.st.reset()
}

func (vm *VM) SetReader(r input.Reader) {
	vm.input = input.NewBufferedReader(r)
}

func (vm *VM) Exec(memtbl memo.Table) (bool, input.Pos, []capt) {
	idata := vm.code.data.Insns

loop:
	for {
		op := idata[vm.ip]
		switch op {
		case opChar:
			b := decodeU8(idata[vm.ip+1:])
			in, ok := vm.input.Peek()
			if ok && b == in {
				vm.input.Advance(1)
				vm.ip += szChar
			} else {
				goto fail
			}
		case opJump, opBigJump, opSmallJump:
			lbl, _ := decodeLabel(idata[vm.ip:], op)
			vm.ip += int(lbl)
		case opChoice, opBigChoice, opSmallChoice:
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
			vm.ip += sz
		case opCall, opBigCall, opSmallCall:
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			vm.st.pushRet(stackRet(vm.ip + sz))
			vm.ip += int(lbl)
		case opCommit, opBigCommit, opSmallCommit:
			lbl, _ := decodeLabel(idata[vm.ip:], op)
			vm.st.pop()
			vm.ip += int(lbl)
		case opReturn:
			ent := vm.st.pop()
			if ent != nil && ent.stype == stRet {
				vm.ip = int(ent.ret)
			} else {
				panic("Return failed")
			}
		case opFail:
			goto fail
		case opSet:
			set := decodeSet(idata[vm.ip+1:], vm.code.data.Sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += szSet
			} else {
				goto fail
			}
		case opAny:
			n := decodeU8(idata[vm.ip+1:])
			ok := vm.input.Advance(int(n))
			if ok {
				vm.ip += szAny
			} else {
				goto fail
			}
		case opPartialCommit, opBigPartialCommit, opSmallPartialCommit:
			lbl, _ := decodeLabel(idata[vm.ip:], op)
			ent := vm.st.peek()
			if ent != nil && ent.stype == stBtrack {
				ent.btrack.off = vm.input.Offset()
				ent.btrack.capt = vm.capt
				vm.ip += int(lbl)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			set := decodeSet(idata[vm.ip+1:], vm.code.data.Sets)
			in, ok := vm.input.Peek()
			for ok && set.Has(in) {
				vm.input.Advance(1)
				in, ok = vm.input.Peek()
			}
			vm.ip += szSpan
		case opBackCommit, opBigBackCommit, opSmallBackCommit:
			lbl, _ := decodeLabel(idata[vm.ip:], op)
			ent := vm.st.pop()
			if ent != nil && ent.stype == stBtrack {
				vm.input.SeekTo(ent.btrack.off)
				vm.capt = ent.btrack.capt
				vm.ip += int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			vm.st.pop()
			goto fail
		case opTestChar, opBigTestChar:
			b := decodeU8(idata[vm.ip+1:])
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += sz
			} else {
				vm.ip += int(lbl)
			}
		case opTestCharNoChoice, opBigTestCharNoChoice:
			b := decodeU8(idata[vm.ip+1:])
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.input.Advance(1)
				vm.ip += szTestCharNoChoice
			} else {
				lbl, _ := decodeLabel(idata[vm.ip:], op)
				vm.ip += int(lbl)
			}
		case opTestSet, opBigTestSet:
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			set := decodeSet(idata[vm.ip+1:], vm.code.data.Sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += sz
			} else {
				vm.ip += int(lbl)
			}
		case opTestAny, opBigTestAny:
			n := decodeU8(idata[vm.ip+1:])
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			ent := stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt}
			ok := vm.input.Advance(int(n))
			if ok {
				vm.st.pushBacktrack(ent)
				vm.ip += sz
			} else {
				vm.ip += int(lbl)
			}
		case opCaptureBegin, opCaptureEnd:
			c := capt{
				ip:  vm.ip,
				off: vm.input.Offset(),
			}
			vm.capt = append(vm.capt, c)
			if op == opCaptureBegin {
				vm.ip += szCaptureBegin
			} else {
				vm.ip += szCaptureEnd
			}
		case opCaptureLate, opCaptureFull:
			back := decodeU8(idata[vm.ip+1:])
			c := capt{
				ip:  vm.ip,
				off: vm.input.Offset() - input.Pos(back),
			}
			vm.capt = append(vm.capt, c)
			vm.ip += szCaptureLate
		case opEnd:
			// ends the machine with a success
			break loop
		case opMemoOpen, opBigMemoOpen:
			lbl, sz := decodeLabel(idata[vm.ip:], op)
			id := decodeI16(idata[vm.ip+2:])

			ment, ok := memtbl.Get(memo.Key{
				Id:  uint16(id),
				Pos: vm.input.Offset(),
			})
			if ok {
				if ment.MatchLength() == -1 {
					goto fail
				}
				vm.input.Advance(ment.MatchLength())
				vm.ip += int(lbl)
			} else {
				vm.st.pushMemo(stackMemo{
					id:  uint16(id),
					pos: vm.input.Offset(),
				})
				vm.ip += sz
			}
		case opMemoClose:
			ent := vm.st.pop()
			if ent != nil && ent.stype == stMemo {
				mlen := int(vm.input.Offset()) - int(ent.memo.pos)
				if mlen >= memoCutoff {
					memtbl.Put(memo.Key{
						Id:  ent.memo.id,
						Pos: ent.memo.pos,
					}, memo.NewEntry(mlen, int(vm.input.MaxExaminedPos())-int(ent.memo.pos)+1, nil)) // TODO: +1?
				}
				vm.ip += szMemoClose
			} else {
				panic("MemoClose found no partial memo entry!")
			}
		default:
			panic("Invalid opcode")
		}
	}

	return true, vm.input.Offset(), vm.capt

fail:
	ent := vm.st.pop()
	if ent == nil {
		// match failed
		return false, vm.input.Offset(), vm.capt
	}

	switch ent.stype {
	case stBtrack:
		vm.ip = ent.btrack.ip
		vm.input.SeekTo(ent.btrack.off)
		vm.capt = ent.btrack.capt
	case stMemo:
		// Mark this position in the memoTable as a failed match
		mlen := int(vm.input.Offset()) - int(ent.memo.pos)
		if mlen >= memoCutoff {
			memtbl.Put(memo.Key{
				Id:  ent.memo.id,
				Pos: ent.memo.pos,
			}, memo.NewEntry(-1, -1, nil))
		}
		goto fail
	case stRet:
		goto fail
	}

	goto loop
}

func decodeU8(b []byte) byte {
	return b[0]
}

func decodeI8(b []byte) int8 {
	return *(*int8)(unsafe.Pointer(&b[0]))
}

func decodeI16(b []byte) int16 {
	return *(*int16)(unsafe.Pointer(&b[0]))
}

func decodeI32(b []byte) int32 {
	i1 := uint32(*(*uint16)(unsafe.Pointer(&b[0])))
	i2 := uint32(*(*uint16)(unsafe.Pointer(&b[2])))
	i := (i1 << 16) | i2
	return int32(i)
}

// Decodes the label in the instruction b, given that the opcode for the
// instruction was op. Returns the value of the label and the size of the
// instruction.
func decodeLabel(b []byte, op byte) (int, int) {
	switch op {
	case opBigJump, opBigChoice, opBigCall, opBigCommit,
		opBigPartialCommit, opBigBackCommit, opBigTestChar,
		opBigTestCharNoChoice, opBigTestSet, opBigTestAny:
		return int(decodeI32(b[2:])), szBigJump
	case opBigMemoOpen:
		return int(decodeI32(b[4:])), szBigMemoOpen
	case opSmallJump, opSmallChoice, opSmallCall, opSmallCommit,
		opSmallPartialCommit, opSmallBackCommit:
		return int(decodeI8(b[1:])), szSmallJump
	case opMemoOpen:
		return int(decodeI16(b[4:])), szMemoOpen
	default:
		return int(decodeI16(b[2:])), szJump
	}
}

func decodeSet(b []byte, sets []charset.Set) charset.Set {
	i := decodeU8(b)
	return sets[i]
}
