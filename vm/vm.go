// Package vm implements the GPeg virtual machine.
package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"regexp/syntax"

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
	if vm.prog == nil {
		vm.predecode()
	}
	prog := vm.prog

	if ip < 0 || ip >= len(prog) {
		return true, 0, memo.NewCaptureDummy(0, 0, nil), nil
	}

	var caprange Interval

	if intrvl != nil {
		caprange = *intrvl
		// Apply an edit that clears all memoized entries in the interval
		// we are capturing. This ensures that we find all captures in the
		// requested interval. An empty interval (used for capture-less
		// memo-filling passes) is skipped: the edit would be a no-op, but
		// recording it stamps a shift that every later memo access pays
		// to examine.
		if intrvl.High > intrvl.Low {
			memtbl.ApplyEdit(memo.Edit{
				Start: intrvl.Low,
				End:   intrvl.High,
				Len:   intrvl.High - intrvl.Low,
			})
		}
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
		in := &prog[ip]
		switch in.op {
		case opChar:
			b, ok := src.Peek()
			if ok && byte(in.a) == b {
				src.Advance(1)
				ip++
			} else {
				goto fail
			}
		case opJump:
			ip = int(in.a)
		case opChoice:
			st.pushBacktrack(stackBacktrack{int(in.a), src.Pos()})
			ip++
		case opCall:
			st.pushRet(stackRet(ip + 1))
			ip = int(in.a)
		case opCommit:
			st.pop(true)
			ip = int(in.a)
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
			b, ok := src.Peek()
			if ok && in.set.Has(b) {
				src.Advance(1)
				ip++
			} else {
				goto fail
			}
		case opAny:
			if src.Advance(int(in.a)) {
				ip++
			} else {
				goto fail
			}
		case opPartialCommit:
			ent := st.peek()
			if ent != nil && ent.stype == stBtrack {
				ent.btrack.off = src.Pos()
				st.propCapt()
				ent.capt = nil
				ip = int(in.a)
			} else {
				panic("PartialCommit failed")
			}
		case opSpan:
			src.Span(in.set)
			ip++
		case opBackCommit:
			ent := st.pop(true)
			if ent != nil && ent.stype == stBtrack {
				src.SeekTo(ent.btrack.off)
				ip = int(in.a)
			} else {
				panic("BackCommit failed")
			}
		case opFailTwice:
			st.pop(false)
			goto fail
		case opEmpty:
			op := syntax.EmptyOp(in.a)
			r1, r2 := rune(-1), rune(-1)
			// TODO: PeekBefore may cause problems with incremental parsing
			b1, ok := src.PeekBefore()
			if ok {
				r1 = rune(b1)
			}
			b2, ok := src.Peek()
			if ok {
				r2 = rune(b2)
			}
			sat := syntax.EmptyOpContext(r1, r2)
			if (sat & op) != 0 {
				ip++
			} else {
				goto fail
			}
		case opTestChar:
			b, ok := src.Peek()
			if ok && b == byte(in.b) {
				st.pushBacktrack(stackBacktrack{int(in.a), src.Pos()})
				src.Advance(1)
				ip++
			} else {
				ip = int(in.a)
			}
		case opTestCharNoChoice:
			b, ok := src.Peek()
			if ok && b == byte(in.b) {
				src.Advance(1)
				ip++
			} else {
				ip = int(in.a)
			}
		case opTestSet:
			b, ok := src.Peek()
			if ok && in.set.Has(b) {
				st.pushBacktrack(stackBacktrack{int(in.a), src.Pos()})
				src.Advance(1)
				ip++
			} else {
				ip = int(in.a)
			}
		case opTestSetNoChoice:
			b, ok := src.Peek()
			if ok && in.set.Has(b) {
				src.Advance(1)
				ip++
			} else {
				ip = int(in.a)
			}
		case opTestAny:
			ent := stackBacktrack{int(in.a), src.Pos()}
			if src.Advance(int(in.b)) {
				st.pushBacktrack(ent)
				ip++
			} else {
				ip = int(in.a)
			}
		case opCaptureBegin:
			st.pushCapt(stackMemo{
				id:  int16(in.a),
				pos: src.Pos(),
			})
			ip++
		case opCaptureLate:
			st.pushCapt(stackMemo{
				id:  int16(in.a),
				pos: src.Pos() - int(in.b),
			})
			ip++
		case opCaptureFull:
			back := int(in.b)
			pos := src.Pos()

			if overlaps(intrvl, pos-back, pos) {
				caprange.Low = min(caprange.Low, pos-back)
				caprange.High = max(caprange.High, pos)
				capt := memo.NewCaptureNode(int(in.a), pos-back, back, nil)
				st.addCapt(capt)
			}

			ip++
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
			ip++
		case opEnd:
			success = in.a != 1
			break loop
		case opMemoOpen:
			ment, ok := memtbl.Get(int(in.b), src.Pos())
			if ok {
				if ment.Length() == -1 {
					goto fail
				}
				capt := ment.Captures()
				if capt != nil {
					st.addCapt(capt...)
				}
				src.Advance(ment.Length())
				ip = int(in.a)
			} else {
				st.pushMemo(stackMemo{
					id:  int16(in.b),
					pos: src.Pos(),
				})
				ip++
			}
		case opMemoClose:
			ent := st.pop(true)
			if ent != nil && ent.stype == stMemo {
				mlen := src.Pos() - ent.memo.pos
				memoize(int(ent.memo.id), ent.memo.pos, mlen, 1, ent.capt)
			} else {
				panic("memo close failed")
			}
			ip++
		case opMemoTreeOpen:
			ment, ok := memtbl.Get(int(in.b), src.Pos())
			if ok {
				if ment.Length() == -1 {
					goto fail
				}
				st.pushMemoTree(stackMemo{
					id:    int16(in.b),
					pos:   src.Pos(),
					count: ment.Count(),
				})
				capt := ment.Captures()
				if capt != nil {
					st.addCapt(capt...)
				}
				src.Advance(ment.Length())
				src.Peek()
				ip = int(in.a)
			} else {
				st.pushMemoTree(stackMemo{
					id:  int16(in.b),
					pos: src.Pos(),
				})
				ip++
			}
		case opMemoTreeClose:
			id := int16(in.a)
			for p := st.peek(); p != nil && p.stype == stMemoTree && p.memo.id == id; p = st.peek() {
				st.pop(true)
			}
			ip++
		case opMemoTreeInsert:
			ent := st.peek()
			if ent == nil || ent.stype != stMemoTree {
				panic("no memo entry on stack")
			}
			mlen := src.Pos() - ent.memo.pos
			ent.memo.count++
			memoize(int(ent.memo.id), ent.memo.pos, mlen, ent.memo.count, ent.capt)
			ip++
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

			ip++
		case opCheckBegin:
			st.pushCheck(stackMemo{
				id:    int16(in.a),
				count: int(in.b),
				pos:   src.Pos(),
			})
			ip++
		case opCheckEnd:
			ent := st.pop(true)
			if ent == nil || ent.stype != stCheck {
				panic("check end needs check stack entry")
			}
			checker := vm.data.Checkers[int(in.a)]

			id := int(ent.memo.id)
			flag := ent.memo.count
			n := checker.Check(src.Slice(int(ent.memo.pos), src.Pos()), src, id, flag)
			if n == -1 {
				goto fail
			} else {
				src.Advance(n)
			}

			ip++
		case opError:
			msg := vm.data.Errors[int(in.a)]
			errs = append(errs, ParseError{
				Pos:     src.Pos(),
				Message: msg,
			})
			ip++
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

// decodeSet returns a pointer into the code's set table: sets are 32
// bytes, and set instructions are hot enough that copying one per
// execution is measurable.
func decodeSet(b []byte, sets []charset.Set) *charset.Set {
	i := decodeU8(b)
	return &sets[i]
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
