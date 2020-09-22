package vm

import "github.com/zyedidia/gpeg/input"

type stack struct {
	entries []stackEntry
}

func newStack() *stack {
	return &stack{
		entries: []stackEntry{},
	}
}

func (s *stack) push(ent stackEntry) {
	s.entries = append(s.entries, ent)
}

func (s *stack) pop() (stackEntry, bool) {
	if len(s.entries) == 0 {
		return stackEntry{}, false
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return ret, true
}

func (s *stack) peek() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}

func (s *stack) backtrack(ip int, off input.Pos) stackEntry {
	return stackEntry{
		retaddr: -1,
		btrack: backtrack{
			ip:  ip,
			off: off,
		},
	}
}

func (s *stack) retaddr(addr int) stackEntry {
	return stackEntry{
		retaddr: addr,
	}
}

type stackEntry struct {
	// if retaddr is -1 use btrack instead
	retaddr int
	btrack  backtrack
}

func (se stackEntry) isRet() bool {
	return se.retaddr != -1
}

type backtrack struct {
	ip  int
	off input.Pos
}
