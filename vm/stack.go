package vm

import (
	"github.com/zyedidia/gpeg/ast"
	"github.com/zyedidia/gpeg/input"
)

type stack struct {
	entries []stackEntry
	capt    []*ast.Node
}

func (s *stack) addCapt(capt ...*ast.Node) {
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
	if len(s.entries) == 1 {
		s.capt = append(s.capt, top.capt...)
	} else {
		s.entries[len(s.entries)-2].addCapt(top.capt)
	}
}

const (
	stRet = iota
	stBtrack
	stMemo
	stCapt
)

type stackEntry struct {
	stype  byte
	ret    stackRet
	btrack stackBacktrack
	memo   stackMemo

	capt []*ast.Node
}

func (se *stackEntry) addCapt(capt []*ast.Node) {
	if se.capt == nil {
		se.capt = capt
	} else {
		se.capt = append(se.capt, capt...)
	}
}

type stackRet int

func (s stackRet) isStackEntry() {}

type stackBacktrack struct {
	ip  int
	off input.Pos
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
		capt:    make([]*ast.Node, 0),
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
	if propagate && ret.capt != nil {
		s.addCapt(ret.capt...)
	}
	return ret
}

func (s *stack) popCapt(end input.Pos) *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	// For capture entries, we create a new node with the corresponding
	// children.
	node := &ast.Node{
		Id:       int16(ret.memo.id),
		Start:    ret.memo.pos,
		End:      end,
		Children: ret.capt,
	}
	s.addCapt(node)
	return &ret
}

func (s *stack) peek() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
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
