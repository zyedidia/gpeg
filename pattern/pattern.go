package pattern

import (
	"strings"

	"github.com/zyedidia/gpeg/charset"
)

// Cap marks a pattern to be captured.
func Cap(p Pattern) Pattern {
	return CapId(p, 0)
}

// CapId marks a pattern with an ID to be captured.
func CapId(p Pattern, id int16) Pattern {
	return &CapNode{
		Patt: p,
		Id:   id,
	}
}

var memoId int16 = 0

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

// NonTerm builds an unresolved non-terminal with a given name.
// NonTerms should be used together with `Grammar` to build a recursive
// grammar.
func NonTerm(name string) Pattern {
	return &NonTermNode{
		Name: name,
	}
}

// CapGrammar is equivalent to grammar but it captures every non-terminal that
// doesn't end with '_' and maps non-terminal names to their capture IDs in the
// map 'nontermIds'.
func CapGrammar(start string, nonterms map[string]Pattern, nontermIds map[string]int16) Pattern {
	var id int16
	for k, v := range nonterms {
		if !strings.HasSuffix(k, "_") {
			nonterms[k] = CapId(v, id)
			nontermIds[k] = id
			id++
		}
	}
	return Grammar(start, nonterms)
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
