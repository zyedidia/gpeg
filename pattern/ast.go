package pattern

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

// A Pattern is an object that can be compiled into a parsing program.
type Pattern interface {
	Compile() (isa.Program, error)
}

// AltNode is the binary operator for alternation.
type AltNode struct {
	Left, Right Pattern
}

// SeqNode is the binary operator for sequences.
type SeqNode struct {
	Left, Right Pattern
}

// StarNode is the operator for the Kleene star.
type StarNode struct {
	Patt Pattern
}

// PlusNode is the operator for the Kleene plus.
type PlusNode struct {
	Patt Pattern
}

// OptionalNode is the operator for making a pattern optional.
type OptionalNode struct {
	Patt Pattern
}

// NotNode is the not predicate.
type NotNode struct {
	Patt Pattern
}

// AndNode is the and predicate.
type AndNode struct {
	Patt Pattern
}

// CapNode marks a pattern to be captured with a certain ID.
type CapNode struct {
	Patt Pattern
	Id   int16
}

// MemoNode marks a pattern to be memoized with a certain ID.
type MemoNode struct {
	Patt Pattern
	Id   int16
}

// GrammarNode represents a grammar of non-terminals and their associated
// patterns. The Grammar must also have an entry non-terminal.
type GrammarNode struct {
	Defs  map[string]Pattern
	Start string
}

// SearchNode represents a search for a certain pattern.
type SearchNode struct {
	Patt Pattern
}

// RepeatNode represents the repetition of a pattern a constant number of
// times.
type RepeatNode struct {
	Patt Pattern
	N    int
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
