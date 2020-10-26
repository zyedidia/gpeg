package vm

import "github.com/zyedidia/gpeg/input"

type stack struct {
	entries []stackEntry
}

// type stackEntry interface {
// 	isStackEntry()
// }

const (
	stRet = iota
	stBtrack
	stMemo
)

type entry struct {
	ret    stackRet
	btrack stackBacktrack
	memo   stackMemo
}

type stackEntry struct {
	stype byte
	u     entryUnion
}

type stackRet int

func (s stackRet) isStackEntry() {}

type stackBacktrack struct {
	ip   int
	off  input.Pos
	capt []capt
}

func (s stackBacktrack) isStackEntry() {}

type stackMemo struct {
	id  uint16
	pos input.Pos
}

func (s stackMemo) isStackEntry() {}

func newStack() *stack {
	return &stack{
		entries: make([]stackEntry, 0, 4),
	}
}

func (s *stack) reset() {
	s.entries = s.entries[:1]
}

func (s *stack) push(ent stackEntry) {
	s.entries = append(s.entries, ent)
}

func (s *stack) pop() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return &ret
}

func (s *stack) peek() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}

func (s *stack) pushRet(r stackRet) {
	var u entryUnion
	u.retPut(r)
	s.push(stackEntry{
		stype: stRet,
		u:     u,
	})
}

func (s *stack) pushBacktrack(b stackBacktrack) {
	var u entryUnion
	u.btrackPut(b)
	s.push(stackEntry{
		stype: stBtrack,
		u:     u,
	})
}

func (s *stack) pushMemo(m stackMemo) {
	var u entryUnion
	u.memoPut(m)
	s.push(stackEntry{
		stype: stMemo,
		u:     u,
	})
}
