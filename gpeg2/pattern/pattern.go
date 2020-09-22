package pattern

import (
	"fmt"
	"log"

	"github.com/zyedidia/gpeg/isa"
)

type openCall struct {
	name string
	isa.Call
}

// A Pattern is a set of instructions that can be used to
// match an input Reader.
type Pattern []isa.Insn

// String returns the string representation of a pattern
func (p Pattern) String() string {
	s := ""
	for i, instr := range p {
		s += fmt.Sprintf("%2d: %v\n", i, instr)
	}
	return s
}

// Literal matches a given string literal.
func Literal(s string) Pattern {
	code := make(Pattern, len(s))
	for i := 0; i < len(s); i++ {
		code[i] = isa.Char{Byte: s[i]}
	}
	return code
}

// Set matches any character in the given set.
func Set(chars isa.Charset) Pattern {
	return Pattern{isa.Set{Chars: chars}}
}

// Any consumes n characters, and only fails if there
// aren't enough input characters left.
func Any(n uint8) Pattern {
	return Pattern{isa.Any{N: n}}
}

// Concat concatenates two patterns: `p1 p2`.
func Concat(p1, p2 Pattern) Pattern {
	code := make(Pattern, 0, len(p1)+len(p2))
	code = append(code, p1...)
	code = append(code, p2...)
	return code
}

// Or returns the ordered choice between two patterns: `p1 / p2`.
func Or(p1, p2 Pattern) Pattern {
	// optimization: if p1 and p2 are charsets, return the union
	if len(p1) == 1 && len(p2) == 1 {
		s1, ok1 := p1[0].(isa.Set)
		s2, ok2 := p2[0].(isa.Set)
		if ok1 && ok2 {
			return Pattern{
				isa.Set{Chars: s1.Chars.Add(s2.Chars)},
			}
		}
	}

	code := make(Pattern, 0, len(p1)+len(p2)+4)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, p1...)
	code = append(code, isa.Commit{Lbl: L2})
	code = append(code, L1)
	code = append(code, p2...)
	code = append(code, L2)
	return code
}

// Star returns the Kleene star repetition of a pattern: `p*`.
// This matches zero or more occurrences of p.
func Star(p Pattern) Pattern {
	// optimization: repeating a charset uses the dedicated instruction 'span'
	if len(p) == 1 {
		s, ok := p[0].(isa.Set)
		if ok {
			return Pattern{
				isa.Span{Chars: s.Chars},
			}
		}
	}

	code := make(Pattern, 0, len(p)+4)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, L1)
	code = append(code, p...)
	code = append(code, isa.PartialCommit{Lbl: L1})
	code = append(code, L2)
	return code
}

// Plus returns the Kleene plus repetition of a pattern: `p+`.
// This matches one or more occurrences of p.
func Plus(p Pattern) Pattern {
	starp := Star(p)
	code := make(Pattern, 0, len(p)+len(starp))
	code = append(code, p...)
	code = append(code, starp...)
	return code
}

// Optional matches at most 1 occurrence of p: `p?`.
func Optional(p Pattern) Pattern {
	return Or(p, Pattern{})
}

// Not returns the not predicate applied to a pattern: `!p`.
// The not predicate succeeds if matching `p` at the current position
// fails, and does not consume any input.
func Not(p Pattern) Pattern {
	code := make(Pattern, 0, len(p)+3)
	L1 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, p...)
	code = append(code, isa.FailTwice{})
	code = append(code, L1)
	return code
}

// And returns the and predicate applied to a pattern: `&p`.
// The and predicate succeeds if matching `p` at the current position
// succeeds and does not consume any input.
// This is equivalent to `!!p`.
func And(p Pattern) Pattern {
	code := make(Pattern, 0, len(p)+7)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	L3 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, p...)
	code = append(code, L2)
	code = append(code, isa.Commit{Lbl: L3})
	code = append(code, L3)
	code = append(code, isa.Fail{})
	code = append(code, L1)
	return code
}

// NonTerm builds an unresolved non-terminal with a given name.
// NonTerms should be used together with `Grammar` to build a recursive
// grammar.
func NonTerm(name string) Pattern {
	return Pattern{
		openCall{name: name},
	}
}

// Grammar builds a grammar from a map of non-terminal patterns.
// Any unresolved non-terminals are resolved with their definitions
// in the map.
func Grammar(start string, nonterms map[string]Pattern) Pattern {
	size := 0
	for _, v := range nonterms {
		size += len(v)
	}

	// total number of instructions is roughly body instructions, labels,
	// return instructions, and the starter code to call the start symbol
	code := make(Pattern, 0, size+2*len(nonterms)+2)
	// add a call for the starting symbol
	code = append(code, openCall{name: start}, isa.End{})

	// place all functions into the code
	labels := make(map[string]isa.Label)
	for k, v := range nonterms {
		label := isa.NewLabel()
		labels[k] = label
		code = append(code, label)
		code = append(code, v...)
		code = append(code, isa.Return{})
	}

	// resolve calls to openCall and do tail call optimization
	for i, insn := range code {
		if oc, ok := insn.(openCall); ok {
			lbl, ok := labels[oc.name]
			if !ok {
				log.Fatal("Undefined non-terminal in grammar:", oc.name)
			}

			// replace this placeholder instruction with a normal call
			var replace isa.Insn = isa.Call{Lbl: lbl}
			// if a call is immediately followed by a return, optimize to
			// a jump for tail call optimization.
			if i+1 < len(code) {
				if _, ok := code[i+1].(isa.Return); ok {
					replace = isa.Jump{Lbl: lbl}
				}
			}

			// perform the replacement of the opencall by either a call or jump
			code[i] = replace
		}
	}

	return code
}
