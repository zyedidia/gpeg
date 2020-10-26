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
	idata := vm.code.insns

loop:
	for {
		op := idata[vm.ip]
		switch op {
		case opChar:
			b := decodeByte(idata[vm.ip+1:])
			in, ok := vm.input.Peek()
			if ok && b == in {
				vm.input.Advance(1)
				vm.ip += 2
			} else {
				goto fail
			}
		case opJump:
			lbl := decodeU16(idata[vm.ip+2:])
			vm.ip += int(lbl)
		case opChoice:
			lbl := decodeU16(idata[vm.ip+2:])
			vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
			vm.ip += 4
		case opCall:
			lbl := decodeU16(idata[vm.ip+2:])
			vm.st.pushRet(stackRet(vm.ip + 4))
			vm.ip += int(lbl)
		case opCommit:
			lbl := decodeU16(idata[vm.ip+2:])
			vm.st.pop()
			vm.ip += int(lbl)
		case opReturn:
			ent := vm.st.pop()
			if ent != nil && ent.stype == stRet {
				vm.ip = int(ent.u.ret())
			} else {
				panic("Return failed")
			}
		case opFail:
			goto fail
		case opSet:
			set := decodeSet(idata[vm.ip+2:], vm.code.sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += 4
			} else {
				goto fail
			}
		case opAny:
			n := decodeByte(idata[vm.ip+1:])
			ok := vm.input.Advance(int(n))
			if ok {
				vm.ip += 2
			} else {
				goto fail
			}
		case opPartialCommit:
			lbl := decodeU16(idata[vm.ip+2:])
			ent := vm.st.peek()
			if ent.stype == stBtrack {
				b := ent.u.btrackPtr()
				b.off = vm.input.Offset()
				b.capt = vm.capt
				vm.ip += int(lbl)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			set := decodeSet(idata[vm.ip+2:], vm.code.sets)
			in, ok := vm.input.Peek()
			for ok && set.Has(in) {
				vm.input.Advance(1)
				in, ok = vm.input.Peek()
			}
			vm.ip += 4
		case opBackCommit:
			lbl := decodeU16(idata[vm.ip+2:])
			ent := vm.st.pop()
			if ent != nil && ent.stype == stBtrack {
				vm.input.SeekTo(ent.u.btrack().off)
				vm.capt = ent.u.btrack().capt
				vm.ip += int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			vm.st.pop()
			goto fail
		case opTestChar:
			b := decodeByte(idata[vm.ip+1:])
			lbl := decodeU16(idata[vm.ip+2:])
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += 4
			} else {
				vm.ip += int(lbl)
			}
		case opTestSet:
			lbl := decodeU16(idata[vm.ip+2:])
			set := decodeSet(idata[vm.ip+2+2:], vm.code.sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.st.pushBacktrack(stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt})
				vm.input.Advance(1)
				vm.ip += 6
			} else {
				vm.ip += int(lbl)
			}
		case opTestAny:
			n := decodeByte(idata[vm.ip+1:])
			lbl := decodeU16(idata[vm.ip+2:])
			ent := stackBacktrack{vm.ip + int(lbl), vm.input.Offset(), vm.capt}
			ok := vm.input.Advance(int(n))
			if ok {
				vm.st.pushBacktrack(ent)
				vm.ip += 4
			} else {
				vm.ip += int(lbl)
			}
		// case opCapture:
		// 	c := capt{
		// 		ip:  vm.ip,
		// 		off: vm.input.Offset(),
		// 	}
		// 	vm.capt = append(vm.capt, c)
		// 	vm.ip += 3
		case opEnd:
			// ends the machine with a success
			break loop
		case opMemoOpen:
			lbl := decodeU16(idata[vm.ip+2:])
			id := decodeU16(idata[vm.ip+2+2:])

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
				vm.ip += 6
			}
		case opMemoClose:
			ent := vm.st.pop()
			if ent != nil && ent.stype == stMemo {
				mlen := int(vm.input.Offset()) - int(ent.u.memo().pos)
				if mlen >= memoCutoff {
					memtbl.Put(memo.Key{
						Id:  ent.u.memo().id,
						Pos: ent.u.memo().pos,
					}, memo.NewEntry(mlen, int(vm.input.MaxExaminedPos())-int(ent.u.memo().pos)+1)) // TODO: +1?
				}
				vm.ip += 2
			} else {
				panic("MemoClose found no partial memo entry!")
			}
		case opNop:
			vm.ip += 2
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
		vm.ip = ent.u.btrack().ip
		vm.input.SeekTo(ent.u.btrack().off)
		vm.capt = ent.u.btrack().capt
	case stMemo:
		// Mark this position in the memoTable as a failed match
		mlen := int(vm.input.Offset()) - int(ent.u.memo().pos)
		if mlen >= memoCutoff {
			memtbl.Put(memo.Key{
				Id:  ent.u.memo().id,
				Pos: ent.u.memo().pos,
			}, memo.NewEntry(-1, -1))
		}
		goto fail
	case stRet:
		goto fail
	}

	goto loop
}

func decodeByte(b []byte) byte {
	return b[0]
}

func decodeU16(b []byte) int16 {
	return *(*int16)(unsafe.Pointer(&b[0]))
}

func decodeSet(b []byte, sets []charset.Set) charset.Set {
	i := decodeU16(b)
	return sets[i]
}
