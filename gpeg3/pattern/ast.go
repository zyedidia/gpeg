package pattern

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

type Pattern interface {
	Compile() (isa.Program, error)
}

type AltNode struct {
	Left  Pattern
	Right Pattern
}

type SeqNode struct {
	Left  Pattern
	Right Pattern
}

type StarNode struct {
	Patt Pattern
}

type PlusNode struct {
	Patt Pattern
}

type OptionalNode struct {
	Patt Pattern
}

type NotNode struct {
	Patt Pattern
}

type AndNode struct {
	Patt Pattern
}

type CapNode struct {
	Patt Pattern
	Id   int16
}

type MemoNode struct {
	Patt Pattern
	Id   int16
}

type GrammarNode struct {
	Defs  map[string]Pattern
	Start string
}

type SearchNode struct {
	Patt Pattern
}

type RepeatNode struct {
	Patt Pattern
	N    int
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
		WalkPattern(t.Left, followInline, fn)
		WalkPattern(t.Right, followInline, fn)
	case *SeqNode:
		WalkPattern(t.Left, followInline, fn)
		WalkPattern(t.Right, followInline, fn)
	case *StarNode:
		WalkPattern(t.Patt, followInline, fn)
	case *PlusNode:
		WalkPattern(t.Patt, followInline, fn)
	case *OptionalNode:
		WalkPattern(t.Patt, followInline, fn)
	case *NotNode:
		WalkPattern(t.Patt, followInline, fn)
	case *AndNode:
		WalkPattern(t.Patt, followInline, fn)
	case *CapNode:
		WalkPattern(t.Patt, followInline, fn)
	case *MemoNode:
		WalkPattern(t.Patt, followInline, fn)
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
