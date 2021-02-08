// Package vm implements the GPeg virtual machine.
package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/zyedidia/gpeg/capture"
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
)

// do not memoize results that are smaller than this threshold.
const memoCutoff = 128

// type CapFunc func(id int16, start input.Pos, size int,
// 	caps []capture.Capture, in *input.Input) capture.Capture

type ParseError struct {
	Message string
	Pos     input.Pos
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%v: %s", e.Pos, e.Message)
}

// A VM is a virtual machine capable of interpreting GPeg programs.
type VM struct {
	ip   int
	st   *stack
	code VMCode

	input *input.Input
}

// NewVM returns a new parsing machine.
func NewVM(r input.ReaderAtPos, code VMCode) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		input: input.NewInput(r),
		code:  code,
	}
}

// SeekTo moves the current subject position to the given position.
func (vm *VM) SeekTo(p input.Pos) {
	vm.input.SeekTo(p)
}

// Reset resets the virtual machine.
func (vm *VM) Reset() {
	vm.ip = 0
	vm.input.ResetFurthest()
	vm.st.reset()
}

// SetReader assigns a new reader to this virtual machine.
func (vm *VM) SetReader(r input.ReaderAtPos) {
	vm.input = input.NewInput(r)
}

// Exec executes the parsing program this virtual machine was created with. It
// returns whether the parse was a match, the last position in the subject
// string that was matched, and any captures that were created.
func (vm *VM) Exec(memtbl memo.Table) (bool, input.Pos, []*capture.Node, []error) {
	idata := vm.code.data.Insns
	success := true
	var errs []error = nil

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
		case opJump:
			lbl := decodeU24(idata[vm.ip+1:])
			vm.ip = int(lbl)
		case opChoice:
			lbl := decodeU24(idata[vm.ip+1:])
			vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Pos()})
			vm.ip += szChoice
		case opCall:
			lbl := decodeU24(idata[vm.ip+1:])
			vm.st.pushRet(stackRet(vm.ip + szCall))
			vm.ip = int(lbl)
		case opCommit:
			lbl := decodeU24(idata[vm.ip+1:])
			vm.st.pop(true)
			vm.ip = int(lbl)
		case opReturn:
			ent := vm.st.pop(true)
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
		case opPartialCommit:
			lbl := decodeU24(idata[vm.ip+1:])
			ent := vm.st.peek()
			if ent != nil && ent.stype == stBtrack {
				ent.btrack.off = vm.input.Pos()
				vm.st.propCapt()
				ent.capt = nil
				vm.ip = int(lbl)
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
		case opBackCommit:
			lbl := decodeU24(idata[vm.ip+1:])
			ent := vm.st.pop(true)
			if ent != nil && ent.stype == stBtrack {
				vm.input.SeekTo(ent.btrack.off)
				vm.ip = int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			vm.st.pop(false)
			goto fail
		case opTestChar:
			b := decodeU8(idata[vm.ip+2:])
			lbl := decodeU24(idata[vm.ip+3:])
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Pos()})
				vm.input.Advance(1)
				vm.ip += szTestChar
			} else {
				vm.ip = int(lbl)
			}
		case opTestCharNoChoice:
			b := decodeU8(idata[vm.ip+2:])
			in, ok := vm.input.Peek()
			if ok && in == b {
				vm.input.Advance(1)
				vm.ip += szTestCharNoChoice
			} else {
				lbl := decodeU24(idata[vm.ip+3:])
				vm.ip = int(lbl)
			}
		case opTestSet:
			lbl := decodeU24(idata[vm.ip+3:])
			set := decodeSet(idata[vm.ip+2:], vm.code.data.Sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Pos()})
				vm.input.Advance(1)
				vm.ip += szTestSet
			} else {
				vm.ip = int(lbl)
			}
		case opTestSetNoChoice:
			set := decodeSet(idata[vm.ip+2:], vm.code.data.Sets)
			in, ok := vm.input.Peek()
			if ok && set.Has(in) {
				vm.input.Advance(1)
				vm.ip += szTestSetNoChoice
			} else {
				lbl := decodeU24(idata[vm.ip+3:])
				vm.ip = int(lbl)
			}
		case opTestAny:
			n := decodeU8(idata[vm.ip+2:])
			lbl := decodeU24(idata[vm.ip+3:])
			ent := stackBacktrack{vm.ip + int(lbl), vm.input.Pos()}
			ok := vm.input.Advance(int(n))
			if ok {
				vm.st.pushBacktrack(ent)
				vm.ip += szTestAny
			} else {
				vm.ip = int(lbl)
			}
		case opCaptureBegin:
			id := decodeI16(idata[vm.ip+2:])
			vm.st.pushCapt(stackMemo{
				id:  id,
				pos: vm.input.Pos(),
			})
			vm.ip += szCaptureBegin
		case opCaptureLate:
			back := decodeU8(idata[vm.ip+1:])
			id := decodeI16(idata[vm.ip+2:])
			vm.st.pushCapt(stackMemo{
				id:  id,
				pos: vm.input.Pos().Move(-int(back)),
			})
			vm.ip += szCaptureLate
		case opCaptureFull:
			back := int(decodeU8(idata[vm.ip+1:]))
			id := decodeI16(idata[vm.ip+2:])
			pos := vm.input.Pos()

			loc := basicLocator{
				start: pos.Move(-back),
				size:  back,
			}
			capt := capture.NewNode(id, loc, nil)
			vm.st.addCapt(capt)

			vm.ip += szCaptureFull
		case opCaptureEnd:
			ent := vm.st.pop(false)

			if ent == nil || ent.stype != stCapt {
				panic("CaptureEnd did not find capture entry")
			}

			end := vm.input.Pos()
			loc := basicLocator{
				start: ent.memo.pos,
				size:  end.Cmp(ent.memo.pos),
			}
			capt := capture.NewNode(ent.memo.id, loc, ent.capt)
			vm.st.addCapt(capt)
			vm.ip += szCaptureEnd
		case opEnd:
			fail := decodeU8(idata[vm.ip+1:])
			success = fail != 1
			break loop
		case opMemoOpen:
			lbl := decodeU24(idata[vm.ip+1:])
			id := decodeI16(idata[vm.ip+4:])

			ment, ok := memtbl.Get(memo.Key{
				Id:  id,
				Pos: vm.input.Pos(),
			})
			if ok {
				if ment.MatchLength() == -1 {
					goto fail
				}
				capt := ment.Value()
				if capt != nil {
					vm.st.addCapt(capt...)
				}
				vm.input.Advance(ment.MatchLength())
				vm.ip = int(lbl)
			} else {
				vm.st.pushMemo(stackMemo{
					id:  id,
					pos: vm.input.Pos(),
				})
				vm.ip += szMemoOpen
			}
		case opMemoClose:
			ent := vm.st.pop(true)
			if ent != nil && ent.stype == stMemo {
				mlen := vm.input.Pos().Cmp(ent.memo.pos)
				if mlen >= memoCutoff {
					memtbl.Put(memo.Key{
						Id:  ent.memo.id,
						Pos: ent.memo.pos,
					}, memo.NewEntry(ent.memo.id, ent.memo.pos, mlen, vm.input.Furthest().Cmp(ent.memo.pos)+1, ent.capt)) // TODO: +1?
				}
				vm.ip += szMemoClose
			} else {
				panic("MemoClose found no partial memo entry!")
			}
		case opError:
			errid := decodeU24(idata[vm.ip+1:])
			msg := vm.code.data.Errors[errid]
			if errs == nil {
				errs = make([]error, 0, 1)
			}
			errs = append(errs, &ParseError{
				Pos:     vm.input.Pos(),
				Message: msg,
			})
			vm.ip += szError
		default:
			panic("Invalid opcode")
		}
	}

	return success, vm.input.Pos(), vm.st.capt, errs

fail:
	ent := vm.st.pop(false)
	if ent == nil {
		// match failed
		return false, vm.input.Pos(), []*capture.Node{}, errs
	}

	switch ent.stype {
	case stBtrack:
		vm.ip = ent.btrack.ip
		vm.input.SeekTo(ent.btrack.off)
		ent.capt = nil
	case stMemo:
		// Mark this position in the memoTable as a failed match
		mlen := vm.input.Pos().Cmp(ent.memo.pos)
		if mlen >= memoCutoff {
			// TODO: support marking invalid intervals
			// memtbl.Put(memo.Key{
			// 	Id:  ent.memo.id,
			// 	Pos: ent.memo.pos,
			// }, memo.NewEntry(ent.memo.id, ent.memo.pos, -1, -1, nil))
		}
		ent.capt = nil
		goto fail
	case stRet:
		ent.capt = nil
		goto fail
	case stCapt:
		ent.capt = nil
		goto fail
	}

	goto loop
}

func decodeU8(b []byte) byte {
	return b[0]
}

func decodeI8(b []byte) int8 {
	return int8(b[0])
}

func decodeU16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b[0:])
}

func decodeI16(b []byte) int16 {
	return int16(binary.LittleEndian.Uint16(b[0:]))
}

func decodeU24(b []byte) uint32 {
	i1 := uint32(decodeU8(b))
	i2 := uint32(decodeU16(b[1:]))
	i := (i1 << 16) | i2
	return i
}

func decodeSet(b []byte, sets []charset.Set) charset.Set {
	i := decodeU8(b)
	return sets[i]
}

type basicLocator struct {
	start input.Pos
	size  int
}

func (l basicLocator) Start() input.Pos {
	return l.start
}

func (l basicLocator) End() input.Pos {
	return l.start.Move(l.size)
}
