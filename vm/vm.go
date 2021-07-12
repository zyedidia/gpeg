// Package vm implements the GPeg virtual machine.
package vm

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
)

type ParseError struct {
	Message string
	Pos     int
}

type Interval struct {
	Low, High int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%v: %s", e.Pos, e.Message)
}

// Exec executes the parsing program this virtual machine was created with. It
// returns whether the parse was a match, the last position in the subject
// string that was matched, and any captures that were created.
func (vm *Code) Exec(r io.ReaderAt, memtbl memo.Table) (bool, int, *memo.Capture, []ParseError) {
	ip := 0
	st := newStack()
	src := input.NewInput(r)

	// parse in parallel?
	// if memtbl.Size() == 0 {
	// 	srccopy := input.NewInput(r)
	// 	srccopy.SeekTo(1000000)
	// 	go vm.exec(0, newStack(), srccopy, memtbl)
	// }

	return vm.exec(ip, st, src, memtbl, nil)
}

func (vm *Code) ExecInterval(r io.ReaderAt, memtbl memo.Table, intrvl *Interval) (bool, int, *memo.Capture, []ParseError) {
	ip := 0
	st := newStack()
	src := input.NewInput(r)

	return vm.exec(ip, st, src, memtbl, intrvl)
}

func (vm *Code) exec(ip int, st *stack, src *input.Input, memtbl memo.Table, intrvl *Interval) (bool, int, *memo.Capture, []ParseError) {
	idata := vm.data.Insns

	if ip < 0 || ip >= len(idata) {
		return true, 0, memo.NewCaptureDummy(0, 0, nil), nil
	}

	var caprange Interval

	if intrvl != nil {
		caprange = *intrvl
		// Apply an edit that clears all memoized entries in the interval
		// we are capturing. This ensures that we find all captures in the
		// requested interval.
		memtbl.ApplyEdit(memo.Edit{
			Start: intrvl.Low,
			End:   intrvl.High,
			Len:   intrvl.High - intrvl.Low,
		})
	}

	memoize := func(id, pos, mlen, count int, capt []*memo.Capture) {
		if intrvl != nil {
			capt = nil
		}
		mexam := max(src.Furthest(), src.Pos()) - pos + 1
		memtbl.Put(id, pos, mlen, mexam, count, capt)
	}

	success := true
	var errs []ParseError = nil

loop:
	for {
		op := idata[ip]
		switch op {
		case opChar:
			b := decodeU8(idata[ip+1:])
			in, ok := src.Peek()
			if ok && b == in {
				src.Advance(1)
				ip += szChar
			} else {
				goto fail
			}
		case opJump:
			lbl := decodeU24(idata[ip+1:])
			ip = int(lbl)
		case opChoice:
			lbl := decodeU24(idata[ip+1:])
			st.pushBacktrack(stackBacktrack{int(lbl), src.Pos()})
			ip += szChoice
		case opCall:
			lbl := decodeU24(idata[ip+1:])
			st.pushRet(stackRet(ip + szCall))
			ip = int(lbl)
		case opCommit:
			lbl := decodeU24(idata[ip+1:])
			st.pop(true)
			ip = int(lbl)
		case opReturn:
			ent := st.pop(true)
			if ent != nil && ent.stype == stRet {
				ip = int(ent.ret)
			} else {
				panic("Return failed")
			}
		case opFail:
			goto fail
		case opSet:
			set := decodeSet(idata[ip+1:], vm.data.Sets)
			in, ok := src.Peek()
			if ok && set.Has(in) {
				src.Advance(1)
				ip += szSet
			} else {
				goto fail
			}
		case opAny:
			n := decodeU8(idata[ip+1:])
			ok := src.Advance(int(n))
			if ok {
				ip += szAny
			} else {
				goto fail
			}
		case opPartialCommit:
			lbl := decodeU24(idata[ip+1:])
			ent := st.peek()
			if ent != nil && ent.stype == stBtrack {
				ent.btrack.off = src.Pos()
				st.propCapt()
				ent.capt = nil
				ip = int(lbl)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			set := decodeSet(idata[ip+1:], vm.data.Sets)
			in, ok := src.Peek()
			for ok && set.Has(in) {
				src.Advance(1)
				in, ok = src.Peek()
			}
			ip += szSpan
		case opBackCommit:
			lbl := decodeU24(idata[ip+1:])
			ent := st.pop(true)
			if ent != nil && ent.stype == stBtrack {
				src.SeekTo(ent.btrack.off)
				ip = int(lbl)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			st.pop(false)
			goto fail
		case opTestChar:
			b := decodeU8(idata[ip+2:])
			lbl := decodeU24(idata[ip+3:])
			in, ok := src.Peek()
			if ok && in == b {
				st.pushBacktrack(stackBacktrack{int(lbl), src.Pos()})
				src.Advance(1)
				ip += szTestChar
			} else {
				ip = int(lbl)
			}
		case opTestCharNoChoice:
			b := decodeU8(idata[ip+2:])
			in, ok := src.Peek()
			if ok && in == b {
				src.Advance(1)
				ip += szTestCharNoChoice
			} else {
				lbl := decodeU24(idata[ip+3:])
				ip = int(lbl)
			}
		case opTestSet:
			lbl := decodeU24(idata[ip+3:])
			set := decodeSet(idata[ip+2:], vm.data.Sets)
			in, ok := src.Peek()
			if ok && set.Has(in) {
				st.pushBacktrack(stackBacktrack{int(lbl), src.Pos()})
				src.Advance(1)
				ip += szTestSet
			} else {
				ip = int(lbl)
			}
		case opTestSetNoChoice:
			set := decodeSet(idata[ip+2:], vm.data.Sets)
			in, ok := src.Peek()
			if ok && set.Has(in) {
				src.Advance(1)
				ip += szTestSetNoChoice
			} else {
				lbl := decodeU24(idata[ip+3:])
				ip = int(lbl)
			}
		case opTestAny:
			n := decodeU8(idata[ip+2:])
			lbl := decodeU24(idata[ip+3:])
			ent := stackBacktrack{int(lbl), src.Pos()}
			ok := src.Advance(int(n))
			if ok {
				st.pushBacktrack(ent)
				ip += szTestAny
			} else {
				ip = int(lbl)
			}
		case opCaptureBegin:
			id := decodeI16(idata[ip+2:])
			st.pushCapt(stackMemo{
				id:  id,
				pos: src.Pos(),
			})
			ip += szCaptureBegin
		case opCaptureLate:
			back := decodeU8(idata[ip+1:])
			id := decodeI16(idata[ip+2:])
			st.pushCapt(stackMemo{
				id:  id,
				pos: src.Pos() - int(back),
			})
			ip += szCaptureLate
		case opCaptureFull:
			back := int(decodeU8(idata[ip+1:]))
			id := decodeI16(idata[ip+2:])
			pos := src.Pos()

			if overlaps(intrvl, pos-back, pos) {
				caprange.Low = min(caprange.Low, pos-back)
				caprange.High = max(caprange.High, pos)
				capt := memo.NewCaptureNode(int(id), pos-back, back, nil)
				st.addCapt(capt)
			}

			ip += szCaptureFull
		case opCaptureEnd:
			ent := st.pop(false)

			if ent == nil || ent.stype != stCapt {
				panic("CaptureEnd did not find capture entry")
			}

			end := src.Pos()
			if overlaps(intrvl, ent.memo.pos, end) {
				caprange.Low = min(caprange.Low, ent.memo.pos)
				caprange.High = max(caprange.High, end)
				capt := memo.NewCaptureNode(int(ent.memo.id), ent.memo.pos, end-ent.memo.pos, ent.capt)
				st.addCapt(capt)
			}
			ip += szCaptureEnd
		case opEnd:
			fail := decodeU8(idata[ip+1:])
			success = fail != 1
			break loop
		case opMemoOpen:
			lbl := decodeU24(idata[ip+1:])
			id := decodeI16(idata[ip+4:])

			ment, ok := memtbl.Get(int(id), src.Pos())
			if ok {
				if ment.Length() == -1 {
					goto fail
				}
				capt := ment.Captures()
				if capt != nil {
					st.addCapt(capt...)
				}
				src.Advance(ment.Length())
				ip = int(lbl)
			} else {
				st.pushMemo(stackMemo{
					id:  id,
					pos: src.Pos(),
				})
				ip += szMemoOpen
			}
		case opMemoClose:
			ent := st.pop(true)
			if ent != nil && ent.stype == stMemo {
				mlen := src.Pos() - ent.memo.pos
				memoize(int(ent.memo.id), ent.memo.pos, mlen, 1, ent.capt)
			} else {
				panic("memo close failed")
			}
			ip += szMemoClose
		case opMemoTreeOpen:
			lbl := decodeU24(idata[ip+1:])
			id := decodeI16(idata[ip+4:])

			ment, ok := memtbl.Get(int(id), src.Pos())
			if ok {
				if ment.Length() == -1 {
					goto fail
				}
				st.pushMemoTree(stackMemo{
					id:    id,
					pos:   src.Pos(),
					count: ment.Count(),
				})
				capt := ment.Captures()
				if capt != nil {
					st.addCapt(capt...)
				}
				src.Advance(ment.Length())
				src.Peek()
				ip = int(lbl)
			} else {
				st.pushMemoTree(stackMemo{
					id:  id,
					pos: src.Pos(),
				})
				ip += szMemoTreeOpen
			}
		case opMemoTreeClose:
			id := decodeI16(idata[ip+2:])
			for p := st.peek(); p != nil && p.stype == stMemoTree && p.memo.id == id; p = st.peek() {
				st.pop(true)
			}
			ip += szMemoTreeClose
		case opMemoTreeInsert:
			ent := st.peek()
			if ent == nil || ent.stype != stMemoTree {
				panic("no memo entry on stack")
			}
			mlen := src.Pos() - ent.memo.pos
			ent.memo.count++
			memoize(int(ent.memo.id), ent.memo.pos, mlen, ent.memo.count, ent.capt)
			ip += szMemoTreeInsert
		case opMemoTree:
			seen := 0
			accum := 0
			for {
				top := st.peekn(seen)
				next := st.peekn(seen + 1)

				if top == nil || next == nil || top.stype != stMemoTree || next.stype != stMemoTree {
					break
				}

				seen++
				accum += top.memo.count

				if accum < next.memo.count {
					continue
				}

				for i := 0; i < seen-1; i++ {
					st.pop(true)
				}
				ent := st.pop(false) // next is now top of stack

				if len(ent.capt) > 0 && intrvl == nil {
					dummy := memo.NewCaptureDummy(ent.memo.pos, src.Pos()-ent.memo.pos, ent.capt)
					st.addCapt(dummy)
				} else if len(ent.capt) > 0 {
					st.addCapt(ent.capt...)
				}

				next.memo.count = accum + next.memo.count
				mlen := src.Pos() - next.memo.pos
				memoize(int(next.memo.id), next.memo.pos, mlen, next.memo.count, next.capt)

				accum = 0
				seen = 0
			}

			ip += szMemoTree
		case opCheckBegin:
			st.pushCheck(stackRet(src.Pos()))
			ip += szCheckBegin
		case opCheckEnd:
			ent := st.pop(true)
			if ent == nil || ent.stype != stCheck {
				panic("check end needs check stack entry")
			}
			checkid := decodeU24(idata[ip+1:])
			checker := vm.data.Checkers[checkid]

			if !checker.Check(src.Slice(int(ent.ret), src.Pos())) {
				goto fail
			}

			ip += szCheckEnd
		case opError:
			errid := decodeU24(idata[ip+1:])
			msg := vm.data.Errors[errid]
			errs = append(errs, ParseError{
				Pos:     src.Pos(),
				Message: msg,
			})
			ip += szError
		default:
			panic("Invalid opcode")
		}
	}

	if intrvl != nil {
		return success, src.Pos(), memo.NewCaptureDummy(caprange.Low, caprange.High-caprange.Low, st.capt), errs
	}
	return success, src.Pos(), memo.NewCaptureDummy(0, src.Pos(), st.capt), errs

fail:
	ent := st.pop(false)
	if ent == nil {
		// match failed
		return false, src.Pos(), nil, errs
	}

	switch ent.stype {
	case stBtrack:
		ip = ent.btrack.ip
		src.SeekTo(ent.btrack.off)
		ent.capt = nil
	case stMemo:
		// Mark this position in the memoTable as a failed match
		memoize(int(ent.memo.id), ent.memo.pos, -1, 0, nil)
		ent.capt = nil
		goto fail
	case stRet, stCapt, stCheck:
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

func overlaps(i *Interval, low2, high2 int) bool {
	if i == nil {
		return true
	}
	return i.Low < high2 && i.High > low2
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
