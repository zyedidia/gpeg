package gpeg

import (
	"fmt"
	"log"
	"unicode/utf8"
)

// A Pattern is a set of instructions that can be used to
// match an input Reader.
type Pattern []instr

// String returns the string representation of a pattern
func (p Pattern) String() string {
	s := ""
	for i, instr := range p {
		s += fmt.Sprintf("%2d: %v\n", i, instr)
	}
	return s
}

// Match returns the number of characters that this pattern
// matched in the input from the beginning.
// Returns -1 if the match failed.
func (p Pattern) Match(r Reader) int {
	vm := NewVM(r)
	return vm.exec(p)
}

// Literal matches a given string literal
func Literal(s string) Pattern {
	code := make([]instr, utf8.RuneCountInString(s))
	for i, r := range s {
		code[i] = iChar{r}
	}
	return code
}

// Set matches any character in the given set.
func Set(chars charset) Pattern {
	return []instr{
		iCharset{chars},
	}
}

// Any consumes n characters, and only fails if there
// aren't enough input characters left.
func Any(n int) Pattern {
	return []instr{
		iAny{n},
	}
}

// Concat concatenates two patterns: `p1 p2`.
func Concat(p1 Pattern, p2 Pattern) Pattern {
	return append(p1, p2...)
}

// Or returns the ordered choice between two patterns: `p1 / p2`.
func Or(p1 Pattern, p2 Pattern) Pattern {
	// optimization: if p1 and p2 are charsets, return the union
	if len(p1) == 1 && len(p2) == 1 {
		s1, ok1 := p1[0].(iCharset)
		s2, ok2 := p2[0].(iCharset)
		if ok1 && ok2 {
			return []instr{
				iCharset{s1.set.Add(s2.set)},
			}
		}
	}

	var firsti instr = iChoice{len(p1) + 2}
	testi, choicei, match := optHeadFail(p1, len(p1)+2)
	if match {
		p1[0] = choicei
		firsti = testi
	}

	code := append([]instr{
		firsti,
	}, p1...)
	code = append(code, iCommit{len(p2) + 1})
	code = append(code, p2...)
	return code
}

// Star returns the Kleene star repetition of a pattern: `p*`.
// This matches zero or more occurrences of p.
func Star(p Pattern) Pattern {
	// optimization: repeating a charset uses the dedicated
	// instruction 'span'
	if len(p) == 1 {
		s, ok := p[0].(iCharset)
		if ok {
			return []instr{
				iSpan{s.set},
			}
		}
	}

	code := append([]instr{
		iChoice{len(p) + 2},
	}, p...)
	code = append(code, iPartialCommit{-len(p)})
	return code
}

// Plus returns the Kleene plus repetition of a pattern: `p+`.
// This matches one or more occurrences of p.
func Plus(p Pattern) Pattern {
	code := append(p, Star(p)...)
	return code
}

// Optional matches at most 1 occurrence of p: `p?`.
func Optional(p Pattern) Pattern {
	return Or(p, []instr{})
}

// Not returns the not predicate applied to a pattern: `!p`.
// The not predicate succeeds if matching `p` at the current position
// fails, and does not consume any input.
func Not(p Pattern) Pattern {
	var firsti instr = iChoice{len(p) + 2}
	testi, choicei, match := optHeadFail(p, len(p)+2)
	if match {
		p[0] = choicei
		firsti = testi
	}

	code := append([]instr{
		firsti,
	}, p...)
	code = append(code, iFailTwice{})
	return code
}

// And returns the and predicate applied to a pattern: `&p`.
// The and predicate succeeds if matching `p` at the current position
// succeeds and does not consume any input.
// This is equivalent to `!!p`.
func And(p Pattern) Pattern {
	var firsti instr = iChoice{len(p) + 2}
	testi, choicei, match := optHeadFail(p, len(p)+2)
	if match {
		p[0] = choicei
		firsti = testi
	}

	code := append([]instr{
		firsti,
	}, p...)
	code = append(code, iBackCommit{2}, iFail{})
	return code
}

// NonTerm builds an unresolved non-terminal with a given name.
// NonTerms should be used together with `Grammar` to build a recursive
// grammar.
func NonTerm(name string) Pattern {
	return []instr{
		iOpenCall{name},
	}
}

// Grammar builds a grammar from a map of non-terminal patterns.
// Any unresolved non-terminals are resolved with their definitions
// in the map.
func Grammar(start string, nonterminals map[string]Pattern) Pattern {
	code := []instr{}

	offsets := make(map[string]int)

	for k, v := range nonterminals {
		offsets[k] = len(code)
		code = append(code, v...)
		code = append(code, iReturn{})
	}

	startoff, ok := offsets[start]
	if !ok {
		log.Fatal("Undefined start non-terminal in grammar:", start)
	}

	for i, insn := range code {
		if oc, ok := insn.(iOpenCall); ok {
			off, ok := offsets[oc.name]
			if !ok {
				log.Fatal("Undefined non-terminal in grammar:", oc.name)
			}

			var replace instr = iCall{off - i}
			// tail call optimization
			if i+1 < len(code) {
				if _, ok := code[i+1].(iReturn); ok {
					replace = iJump{off - i}
				}
			}
			code[i] = replace
		}
	}

	code = append([]instr{
		iCall{startoff + 2},
		iJump{len(code) + 1},
	}, code...)

	return code
}

// Applies head-fail optimizations to patterns. Returns the corresponding
// TestXXX and Choice2 instructions, and an indicator that the input pattern
// is amenable to the head-fail optimization. The `chLabel` input should be
// the label that the TestXXX instruction and subsequent Choice2 instruction
// should jump to if the test fails.
func optHeadFail(p Pattern, chLabel int) (instr, instr, bool) {
	var testi instr
	var choicei instr

	match := false
	if len(p) >= 1 {
		match = true
		// TODO: unicode for choicei?
		switch t := p[0].(type) {
		case iChar:
			testi = iTestChar{
				x:     t.n,
				label: chLabel,
			}
			choicei = iChoice2{
				label: chLabel - 1,
				back:  1,
			}
		case iCharset:
			testi = iTestCharset{
				set:   t.set,
				label: chLabel,
			}
			choicei = iChoice2{
				label: chLabel - 1,
				back:  1,
			}
		case iAny:
			testi = iTestAny{
				n:     t.n,
				label: chLabel,
			}
			choicei = iChoice2{
				label: chLabel - 1,
				back:  t.n,
			}
		default:
			match = false
		}
	}
	return testi, choicei, match
}
