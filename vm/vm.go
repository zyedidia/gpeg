package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/zyedidia/gpeg/ast"
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
)

// do not memoize results that are smaller than this threshold.
const memoCutoff = 128

// A CapFunc is a function that is called immediately when a capture happens.
// If 'false' is returned then the pattern fails immediately. Note that this
// function may also modify the capture.
type CapFunc func(capt *ast.Node, in *input.BufferedReader) bool

type ParseError struct {
	Message string
	Pos     input.Pos
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%v: %s", e.Pos, e.Message)
}

// A VM is a virtual machine capable of interpreting GPeg programs.
type VM struct {
	ip     int
	st     *stack
	code   VMCode
	capfns map[int16]CapFunc

	input *input.BufferedReader
}

// NewVM returns a new parsing machine.
func NewVM(r input.Reader, code VMCode) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		input: input.NewBufferedReader(r),
		code:  code,
	}
}

// AddCapFunc registers a given capture function to the specified capture ID.
func (vm *VM) AddCapFunc(id int16, fn CapFunc) {
	if vm.capfns == nil {
		vm.capfns = make(map[int16]CapFunc)
	}
	vm.capfns[id] = fn
}

func (vm *VM) callCapFunc(node *ast.Node) bool {
	if vm.capfns == nil {
		return true
	}
	if fn, ok := vm.capfns[node.Id]; ok {
		return fn(node, vm.input)
	}
	return true
}

func (vm *VM) addCapt(nodes ...*ast.Node) bool {
	for _, n := range nodes {
		ok := vm.callCapFunc(n)
		if !ok {
			return false
		}
	}
	vm.st.addCapt(nodes...)
	return true
}

// SeekTo moves the current subject position to the given position.
func (vm *VM) SeekTo(p input.Pos) {
	vm.input.SeekTo(p)
}

// Reset resets the virtual machine.
func (vm *VM) Reset() {
	vm.ip = 0
	vm.input.ResetMaxExamined()
	vm.st.reset()
}

// SetReader assigns a new reader to this virtual machine.
func (vm *VM) SetReader(r input.Reader) {
	vm.input = input.NewBufferedReader(r)
}

// Exec executes the parsing program this virtual machine was created with. It
// returns whether the parse was a match, the last position in the subject
// string that was matched, and any captures that were created.
func (vm *VM) Exec(memtbl memo.Table) (bool, input.Pos, []*ast.Node, []error) {
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
			vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Offset()})
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
				ent.btrack.off = vm.input.Offset()
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
				vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Offset()})
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
				vm.st.pushBacktrack(stackBacktrack{int(lbl), vm.input.Offset()})
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
			ent := stackBacktrack{vm.ip + int(lbl), vm.input.Offset()}
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
				pos: vm.input.Offset(),
			})
			vm.ip += szCaptureBegin
		case opCaptureLate:
			back := decodeU8(idata[vm.ip+1:])
			id := decodeI16(idata[vm.ip+2:])
			vm.st.pushCapt(stackMemo{
				id:  id,
				pos: vm.input.Offset() - input.Pos(back),
			})
			vm.ip += szCaptureLate
		case opCaptureFull:
			back := decodeU8(idata[vm.ip+1:])
			id := decodeI16(idata[vm.ip+2:])
			pos := vm.input.Offset()
			node := ast.NewNode(id, pos-input.Pos(back), int(back), nil)

			success := vm.addCapt(node)
			if !success {
				goto fail
			}

			vm.ip += szCaptureFull
		case opCaptureEnd:
			ent := vm.st.pop(false)
			end := vm.input.Offset()
			node := ast.NewNode(ent.memo.id, ent.memo.pos, int(end-ent.memo.pos), ent.capt)
			success := vm.addCapt(node)
			if !success {
				goto fail
			}
			if ent == nil || ent.stype != stCapt {
				panic("CaptureEnd did not find capture entry")
			}
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
				Pos: vm.input.Offset(),
			})
			if ok {
				if ment.MatchLength() == -1 {
					goto fail
				}
				capt := ment.Value()
				if capt != nil {
					success := vm.addCapt(capt...)
					if !success {
						goto fail
					}
				}
				vm.input.Advance(ment.MatchLength())
				vm.ip = int(lbl)
			} else {
				vm.st.pushMemo(stackMemo{
					id:  id,
					pos: vm.input.Offset(),
				})
				vm.ip += szMemoOpen
			}
		case opMemoClose:
			ent := vm.st.pop(true)
			if ent != nil && ent.stype == stMemo {
				mlen := int(vm.input.Offset()) - int(ent.memo.pos)
				if mlen >= memoCutoff {
					memtbl.Put(memo.Key{
						Id:  ent.memo.id,
						Pos: ent.memo.pos,
					}, memo.NewEntry(mlen, int(vm.input.MaxExaminedPos())-int(ent.memo.pos)+1, ent.capt)) // TODO: +1?
				}
				vm.ip += szMemoClose
			} else {
				panic("MemoClose found no partial memo entry!")
			}
		case opError:
			errid := decodeU16(idata[vm.ip+2:])
			msg := vm.code.data.Errors[errid]
			if errs == nil {
				errs = make([]error, 0, 1)
			}
			errs = append(errs, &ParseError{
				Pos:     vm.input.Offset(),
				Message: msg,
			})
			vm.ip += szError
		default:
			panic("Invalid opcode")
		}
	}

	return success, vm.input.Offset(), vm.st.capt, errs

fail:
	ent := vm.st.pop(false)
	if ent == nil {
		// match failed
		return false, vm.input.Offset(), []*ast.Node{}, errs
	}

	switch ent.stype {
	case stBtrack:
		vm.ip = ent.btrack.ip
		vm.input.SeekTo(ent.btrack.off)
		ent.capt = nil
	case stMemo:
		// Mark this position in the memoTable as a failed match
		mlen := int(vm.input.Offset()) - int(ent.memo.pos)
		if mlen >= memoCutoff {
			memtbl.Put(memo.Key{
				Id:  ent.memo.id,
				Pos: ent.memo.pos,
			}, memo.NewEntry(-1, -1, nil))
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
