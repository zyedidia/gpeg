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
	WalkPattern(p, func(sub Pattern) {
		count++
	})
	return count
}

func WalkPattern(p Pattern, fn WalkFunc) {
	fn(p)
	switch t := p.(type) {
	case *AltNode:
		WalkPattern(t.Left, fn)
		WalkPattern(t.Right, fn)
	case *SeqNode:
		WalkPattern(t.Left, fn)
		WalkPattern(t.Right, fn)
	case *StarNode:
		WalkPattern(t.Patt, fn)
	case *PlusNode:
		WalkPattern(t.Patt, fn)
	case *OptionalNode:
		WalkPattern(t.Patt, fn)
	case *NotNode:
		WalkPattern(t.Patt, fn)
	case *AndNode:
		WalkPattern(t.Patt, fn)
	case *CapNode:
		WalkPattern(t.Patt, fn)
	case *MemoNode:
		WalkPattern(t.Patt, fn)
	case *GrammarNode:
		for _, p := range t.Defs {
			WalkPattern(p, fn)
		}
	case *NonTermNode:
		if t.inlined != nil {
			WalkPattern(t.inlined, fn)
		}
	}
}
