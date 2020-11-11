package pattern

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

type Pattern interface {
	Compile() (isa.Program, error)
}

type UnaryOp struct {
	patt Pattern
}

func (o *UnaryOp) Patt() Pattern {
	switch t := o.patt.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	}
	return o.patt
}

type BinaryOp struct {
	left  Pattern
	right Pattern
}

func (o *BinaryOp) Left() Pattern {
	switch t := o.left.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	}
	return o.left
}
func (o *BinaryOp) Right() Pattern {
	switch t := o.right.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	}
	return o.right
}

type AltNode struct {
	BinaryOp
}

type SeqNode struct {
	BinaryOp
}

type StarNode struct {
	UnaryOp
}

type PlusNode struct {
	UnaryOp
}

type OptionalNode struct {
	UnaryOp
}

type NotNode struct {
	UnaryOp
}

type AndNode struct {
	UnaryOp
}

type CapNode struct {
	UnaryOp
	Id int16
}

type MemoNode struct {
	UnaryOp
	Id int16
}

type GrammarNode struct {
	Defs  map[string]Pattern
	Start string
}

type SearchNode struct {
	UnaryOp
}

type RepeatNode struct {
	UnaryOp
	N int
}

type ClassNode struct {
	Chars charset.Set
}

type LiteralNode struct {
	Str string
}

type NonTermNode struct {
	Name    string
	inlined Pattern
}

type DotNode struct {
	N uint8
}

type EmptyNode struct {
}

type WalkFunc func(sub Pattern)

func CountSubPatterns(p Pattern) int {
	count := 0
	WalkPattern(p, true, func(sub Pattern) {
		count++
	})
	return count
}

func WalkPattern(p Pattern, followInline bool, fn WalkFunc) {
	fn(p)
	switch t := p.(type) {
	case *AltNode:
		WalkPattern(t.left, followInline, fn)
		WalkPattern(t.right, followInline, fn)
	case *SeqNode:
		WalkPattern(t.left, followInline, fn)
		WalkPattern(t.right, followInline, fn)
	case *StarNode:
		WalkPattern(t.patt, followInline, fn)
	case *PlusNode:
		WalkPattern(t.patt, followInline, fn)
	case *OptionalNode:
		WalkPattern(t.patt, followInline, fn)
	case *NotNode:
		WalkPattern(t.patt, followInline, fn)
	case *AndNode:
		WalkPattern(t.patt, followInline, fn)
	case *CapNode:
		WalkPattern(t.patt, followInline, fn)
	case *MemoNode:
		WalkPattern(t.patt, followInline, fn)
	case *GrammarNode:
		for _, p := range t.Defs {
			WalkPattern(p, followInline, fn)
		}
	case *NonTermNode:
		if t.inlined != nil && followInline {
			WalkPattern(t.inlined, followInline, fn)
		}
	}
}
