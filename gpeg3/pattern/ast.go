package pattern

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

// A Pattern is an object that can be compiled into a parsing program.
type Pattern interface {
	Compile() (isa.Program, error)
}

// A UnaryOp represents an operator that applies to one subpattern.
type UnaryOp struct {
	patt Pattern
}

// Patt returns the pattern that this UnaryOp modifies.
func (o *UnaryOp) Patt() Pattern {
	switch t := o.patt.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	case *AltNode:
		set, ok := combine(t.Left(), t.Right())
		if ok {
			return &ClassNode{Chars: set}
		}
	}
	return o.patt
}

// A BinaryOp represents an operator that acts on two subpatterns.
type BinaryOp struct {
	left  Pattern
	right Pattern
}

// Left returns this BinaryOp's left subpattern.
func (o *BinaryOp) Left() Pattern {
	switch t := o.left.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	case *AltNode:
		set, ok := combine(t.Left(), t.Right())
		if ok {
			return &ClassNode{Chars: set}
		}
	}
	return o.left
}

// Right returns this BinaryOp's right subpattern.
func (o *BinaryOp) Right() Pattern {
	switch t := o.right.(type) {
	case *NonTermNode:
		if t.inlined != nil {
			return t.inlined
		}
	case *AltNode:
		set, ok := combine(t.Left(), t.Right())
		if ok {
			return &ClassNode{Chars: set}
		}
	}
	return o.right
}

// AltNode is the binary operator for alternation.
type AltNode struct {
	BinaryOp
}

// SeqNode is the binary operator for sequences.
type SeqNode struct {
	BinaryOp
}

// StarNode is the operator for the Kleene star.
type StarNode struct {
	UnaryOp
}

// PlusNode is the operator for the Kleene plus.
type PlusNode struct {
	UnaryOp
}

// OptionalNode is the operator for making a pattern optional.
type OptionalNode struct {
	UnaryOp
}

// NotNode is the not predicate.
type NotNode struct {
	UnaryOp
}

// AndNode is the and predicate.
type AndNode struct {
	UnaryOp
}

// CapNode marks a pattern to be captured with a certain ID.
type CapNode struct {
	UnaryOp
	Id int16
}

// MemoNode marks a pattern to be memoized with a certain ID.
type MemoNode struct {
	UnaryOp
	Id int16
}

// GrammarNode represents a grammar of non-terminals and their associated
// patterns. The Grammar must also have an entry non-terminal.
type GrammarNode struct {
	Defs  map[string]Pattern
	Start string
}

// SearchNode represents a search for a certain pattern.
type SearchNode struct {
	UnaryOp
}

// RepeatNode represents the repetition of a pattern a constant number of
// times.
type RepeatNode struct {
	UnaryOp
	N int
}

// ClassNode represents a character set.
type ClassNode struct {
	Chars charset.Set
}

// LiteralNode represents a literal string.
type LiteralNode struct {
	Str string
}

// NonTermNode represents the use of a non-terminal. If this non-terminal is
// inlined during compilation, the `inlined` field will point to the pattern
// that is inlined.
type NonTermNode struct {
	Name    string
	inlined Pattern
}

// DotNode represents the pattern to match any byte.
type DotNode struct {
	N uint8
}

// EmtpyNode represents the empty pattern.
type EmptyNode struct {
}

// WalkFunc is a function that takes a pattern.
type WalkFunc func(sub Pattern)

// CountSubPatterns returns the number of subpatterns that exist in the given
// pattern.
func CountSubPatterns(p Pattern) int {
	count := 0
	WalkPattern(p, true, func(sub Pattern) {
		count++
	})
	return count
}

// WalkPattern calls fn for every subpattern contained in p. If followInline
// is true, WalkPattern will walk over inlined patterns as well.
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
