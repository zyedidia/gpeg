package vm

import "github.com/zyedidia/gpeg/input"

type memoEntry struct {
	matchLength int
	maxExamined int
}

type memoKey struct {
	id  uint16
	pos input.Pos
}

type memoTable map[memoKey]memoEntry
