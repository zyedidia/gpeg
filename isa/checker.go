package isa

import (
	"github.com/zyedidia/gpeg/input"
)

// A Checker is used so the user can perform additional custom validation of
// parse results. For example, you might want to parse only 8-bit integers by
// matching [0-9]+ and then using a checker to ensure the matched integer is in
// the range 0-256.
type Checker interface {
	Check(b []byte, src *input.Input, id, flag int) int
}

type MapChecker map[string]struct{}

func NewMapChecker(strs []string) MapChecker {
	m := make(map[string]struct{})
	for _, s := range strs {
		m[s] = struct{}{}
	}
	return m
}

func (m MapChecker) Check(b []byte, src *input.Input, id, flag int) int {
	if _, ok := m[string(b)]; ok {
		return 0
	}
	return -1
}

type RefKind uint8

const (
	RefDef RefKind = iota
	RefUse
	RefBlock
)

type BackReference struct {
	symbols map[int]string
}

func NewBackRef() *BackReference {
	return &BackReference{
		symbols: make(map[int]string),
	}
}

func (r *BackReference) Check(b []byte, src *input.Input, id, flag int) int {
	switch RefKind(flag) {
	case RefDef:
		r.symbols[id] = string(b)
		return 0
	case RefUse:
		back := r.symbols[id]
		buf := make([]byte, len(back))
		n, _ := src.ReadAt(buf, int64(src.Pos()))
		if n == len(buf) && string(buf) == back {
			return n
		}
		return -1
	case RefBlock:
	}
	return 0
}
