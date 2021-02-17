package vm

import (
	"github.com/zyedidia/gpeg/memo"
)

type stack struct {
	entries []stackEntry
	capt    []*memo.Capture
}

func (s *stack) addCapt(capt ...*memo.Capture) {
	if len(s.entries) == 0 {
		s.capt = append(s.capt, capt...)
	} else {
		s.entries[len(s.entries)-1].addCapt(capt)
	}
}

func (s *stack) propCapt() {
	if len(s.entries) == 0 {
		return
	}

	top := s.entries[len(s.entries)-1]
	if top.capt != nil && len(top.capt) > 0 {
		if len(s.entries) == 1 {
			s.capt = append(s.capt, top.capt...)
		} else {
			s.entries[len(s.entries)-2].addCapt(top.capt)
		}
	}
}

const (
	stRet = iota
	stBtrack
	stMemo
	stCapt
	stCheck
)

type stackEntry struct {
	stype byte
	// we could use a union to avoid the space cost but I have found this
	// doesn't impact performance and the space cost itself is quite small
	// because the stack is usually small.
	ret    stackRet // stackRet is reused for stCheck
	btrack stackBacktrack
	memo   stackMemo // stackMemo is reused for stCapt

	capt []*memo.Capture
}

func (se *stackEntry) addCapt(capt []*memo.Capture) {
	if se.capt == nil {
		se.capt = capt
	} else {
		se.capt = append(se.capt, capt...)
	}
}

type stackRet int

type stackBacktrack struct {
	ip  int
	off int
}

type stackMemo struct {
	id    int16
	pos   int
	count int
}

func newStack() *stack {
	return &stack{
		entries: make([]stackEntry, 0, 4),
		capt:    make([]*memo.Capture, 0),
	}
}

func (s *stack) reset() {
	s.capt = nil
	// need to complete remake the slice so that the underlying captures can be
	// released to the garbage collector if the user has no references to them
	// (unused stack entries shouldn't keep references to those captures).
	s.entries = make([]stackEntry, 0, 4)
}

func (s *stack) push(ent stackEntry) {
	s.entries = append(s.entries, ent)
}

// propagate marks whether captures should be propagated up the stack.
func (s *stack) pop(propagate bool) *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}

	ret := &s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	// For non-capture entries, propagate the captures upward.
	// For capture entries, we create a new node with the corresponding
	// children, and this is manually handled by the caller.
	if propagate && ret.capt != nil {
		s.addCapt(ret.capt...)
	}
	return ret
}

func (s *stack) peek() *stackEntry {
	return s.peekn(0)
}

func (s *stack) peekn(n int) *stackEntry {
	if len(s.entries) <= n {
		return nil
	}
	return &s.entries[len(s.entries)-n-1]
}

func (s *stack) pushRet(r stackRet) {
	s.push(stackEntry{
		stype: stRet,
		ret:   r,
	})
}

func (s *stack) pushBacktrack(b stackBacktrack) {
	s.push(stackEntry{
		stype:  stBtrack,
		btrack: b,
	})
}

func (s *stack) pushMemo(m stackMemo) {
	s.push(stackEntry{
		stype: stMemo,
		memo:  m,
	})
}

func (s *stack) pushCapt(m stackMemo) {
	s.push(stackEntry{
		stype: stCapt,
		memo:  m,
	})
}

func (s *stack) pushCheck(r stackRet) {
	s.push(stackEntry{
		stype: stCheck,
		ret:   r,
	})
}
