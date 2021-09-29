// Package pattern provides data types and functions for compiling patterns
// into GPeg VM programs.
package pattern

import (
	"regexp/syntax"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

// Cap marks a pattern to be captured.
func Cap(p Pattern, id int) Pattern {
	return &CapNode{
		Patt: p,
		Id:   id,
	}
}

// Check marks a pattern to be checked with the given checker.
func Check(p Pattern, c isa.Checker) Pattern {
	return &CheckNode{
		Patt:    p,
		Checker: c,
	}
}

var memoId = 0

// MemoId marks a pattern as memoizable with a particular ID.
func MemoId(p Pattern, id int) Pattern {
	m := &MemoNode{
		Patt: p,
		Id:   id,
	}
	memoId = max(memoId, id) + 1
	return m
}

// Memo marks a pattern as memoizable.
func Memo(p Pattern) Pattern {
	m := &MemoNode{
		Patt: p,
		Id:   memoId,
	}
	memoId++
	return m
}

// Literal matches a given string literal.
func Literal(s string) Pattern {
	return &LiteralNode{
		Str: s,
	}
}

// Set matches any character in the given set.
func Set(chars charset.Set) Pattern {
	return &ClassNode{
		Chars: chars,
	}
}

// Any consumes n characters, and only fails if there
// aren't enough input characters left.
func Any(n uint8) Pattern {
	return &DotNode{
		N: n,
	}
}

// Repeat matches p exactly n times
func Repeat(p Pattern, n int) Pattern {
	if n <= 0 {
		return &EmptyNode{}
	}

	acc := p
	for i := 1; i < n; i++ {
		acc = &SeqNode{
			Left:  acc,
			Right: p,
		}
	}
	return acc
}

// Concat concatenates n patterns: `p1 p2 p3...`.
func Concat(patts ...Pattern) Pattern {
	if len(patts) <= 0 {
		return &EmptyNode{}
	}

	acc := patts[0]
	for _, p := range patts[1:] {
		acc = &SeqNode{
			Left:  acc,
			Right: p,
		}
	}

	return acc
}

// Or returns the ordered choice between n patterns: `p1 / p2 / p3...`.
func Or(patts ...Pattern) Pattern {
	if len(patts) <= 0 {
		return &EmptyNode{}
	}

	// optimization: make or right associative
	acc := patts[len(patts)-1]
	for i := len(patts) - 2; i >= 0; i-- {
		acc = &AltNode{
			Left:  patts[i],
			Right: acc,
		}
	}

	return acc
}

// Star returns the Kleene star repetition of a pattern: `p*`.
// This matches zero or more occurrences of p.
func Star(p Pattern) Pattern {
	return &StarNode{
		Patt: p,
	}
}

// Plus returns the Kleene plus repetition of a pattern: `p+`.
// This matches one or more occurrences of p.
func Plus(p Pattern) Pattern {
	return &PlusNode{
		Patt: p,
	}
}

// Optional matches at most 1 occurrence of p: `p?`.
func Optional(p Pattern) Pattern {
	return &OptionalNode{
		Patt: p,
	}
}

// Not returns the not predicate applied to a pattern: `!p`.
// The not predicate succeeds if matching `p` at the current position
// fails, and does not consume any input.
func Not(p Pattern) Pattern {
	return &NotNode{
		Patt: p,
	}
}

// And returns the and predicate applied to a pattern: `&p`.
// The and predicate succeeds if matching `p` at the current position
// succeeds and does not consume any input.
// This is equivalent to `!!p`.
func And(p Pattern) Pattern {
	return &AndNode{
		Patt: p,
	}
}

// Search is a dedicated operator for creating searches. It will match
// the first occurrence of the given pattern. Use Star(Search(p)) to match
// the last occurrence (for a non-overlapping pattern).
func Search(p Pattern) Pattern {
	return &SearchNode{
		Patt: p,
	}
}

func EmptyOp(op syntax.EmptyOp) Pattern {
	return &EmptyOpNode{
		Op: op,
	}
}

// NonTerm builds an unresolved non-terminal with a given name.
// NonTerms should be used together with `Grammar` to build a recursive
// grammar.
func NonTerm(name string) Pattern {
	return &NonTermNode{
		Name: name,
	}
}

// Grammar builds a grammar from a map of non-terminal patterns.
// Any unresolved non-terminals are resolved with their definitions
// in the map.
func Grammar(start string, nonterms map[string]Pattern) Pattern {
	return &GrammarNode{
		Defs:  nonterms,
		Start: start,
	}
}

func CapGrammar(start string, nonterms map[string]Pattern) Pattern {
	m := make(map[string]Pattern)
	id := 0
	for k, v := range nonterms {
		m[k] = Cap(v, id)
		id++
	}
	return Grammar(start, m)
}

// Error is a pattern that throws an error with the given message.
func Error(msg string, recovery Pattern) Pattern {
	return &ErrorNode{
		Message: msg,
		Recover: recovery,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
