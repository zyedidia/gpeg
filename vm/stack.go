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

	top := &s.entries[len(s.entries)-1]
	if len(top.capt) > 0 {
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
	stMemoTree
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
	if len(capt) == 0 {
		return
	}
	if len(se.capt) == 0 {
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

// grow extends the stack by one entry in place and returns a pointer to
// it. Pushes write only the fields their entry type uses (plus a capt
// reset, since the slot may be reused): a full stackEntry literal would
// zero and copy ~80 bytes per push, which shows up heavily on weaker
// machines (duffcopy/duffzero), and pushes are the hottest stack
// operation. Readers always dispatch on stype before touching a field, so
// stale unrelated fields in a reused slot are never observed.
func (s *stack) grow() *stackEntry {
	if len(s.entries) == cap(s.entries) {
		s.entries = append(s.entries, stackEntry{})
	} else {
		s.entries = s.entries[:len(s.entries)+1]
	}
	return &s.entries[len(s.entries)-1]
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
	e := s.grow()
	e.stype = stRet
	e.ret = r
	e.capt = nil
}

func (s *stack) pushBacktrack(b stackBacktrack) {
	e := s.grow()
	e.stype = stBtrack
	e.btrack = b
	e.capt = nil
}

func (s *stack) pushMemo(m stackMemo) {
	e := s.grow()
	e.stype = stMemo
	e.memo = m
	e.capt = nil
}

func (s *stack) pushMemoTree(m stackMemo) {
	e := s.grow()
	e.stype = stMemoTree
	e.memo = m
	e.capt = nil
}

func (s *stack) pushCapt(m stackMemo) {
	e := s.grow()
	e.stype = stCapt
	e.memo = m
	e.capt = nil
}

func (s *stack) pushCheck(m stackMemo) {
	e := s.grow()
	e.stype = stCheck
	e.memo = m
	e.capt = nil
}
