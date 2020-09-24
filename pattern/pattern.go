package pattern

import (
	"fmt"
	"log"

	"github.com/zyedidia/gpeg/isa"
)

// openCall is a dummy instruction for resolving recursive function calls in
// grammars.
type openCall struct {
	name string
	isa.Nop
}

// A Pattern is a set of instructions that can be used to match an input
// Reader.
type Pattern []isa.Insn

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

	code := make(Pattern, 0, len(p1)+len(p2)+5)
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
	starp := Star(p.Copy())
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
	code := make(Pattern, 0, len(p)+4)
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
	code := make(Pattern, 0, len(p)+5)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()

	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, p...)
	code = append(code, isa.BackCommit{Lbl: L2})
	code = append(code, L1)
	code = append(code, isa.Fail{})
	code = append(code, L2)
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
	// this slice may need to expand itself if we do a lot of inlining
	code := make(Pattern, 0, size+2*len(nonterms)+2)
	// add a call for the starting symbol
	LEnd := isa.NewLabel()
	code = append(code, openCall{name: start}, isa.Jump{Lbl: LEnd})

	inlinable := make(map[string]bool)

	// TODO: more aggressive inlining
	for k, v := range nonterms {
		inlinable[k] = true
		for i, insn := range v {
			if oc, ok := insn.(openCall); ok {
				// If we see an openCall this generally means the function is
				// not inlinable. The only exception is a recursive tail-call.
				// It's possible that more aggressive inlining than this could
				// be done.
				_, ok := nextInsn(v[i+1:])
				// recursive tail-call is ok to inline
				inlinable[k] = oc.name == k && !ok
				if !inlinable[k] {
					break
				}
			}
		}

		if inlinable[k] {
			// if this function is inlinable we need to copy it because we
			// will modify the inline code
			fn := v.Copy()
			nonterms[k] = fn
		}
	}

	// place all non-inlined functions into the code
	labels := make(map[string]isa.Label)
	for k, v := range nonterms {
		if inlinable[k] {
			continue
		}
		label := isa.NewLabel()
		labels[k] = label
		code = append(code, label)
		code = append(code, v...)
		code = append(code, isa.Return{})
	}

	// do tail-call optimization for inlinable functions
	for k, fn := range nonterms {
		if !inlinable[k] {
			continue
		}
		for i, insn := range fn {
			if oc, ok := insn.(openCall); ok {
				if oc.name == k {
					// recursive tail-call
					label := isa.NewLabel()
					fn[i] = isa.Jump{Lbl: label}
					nonterms[k] = append(Pattern{label}, fn...)
					break
				}

				// technically we have support for optimizing tail-calls to
				// non-inlinable functions but this should never happens
				// because this would cause this current function to become
				// non-inlinable
				lbl, ok := labels[oc.name]
				if ok {
					fn[i] = isa.Jump{Lbl: lbl}
					log.Println("???: Tail-call to non-inlinable function")
				} else {
					log.Fatal("Error: inlinable function has non-recursive openCall")
				}
			}
		}
	}

	// resolve calls to openCall and do tail call optimization
	for i := 0; i < len(code); i++ {
		insn := code[i]
		if oc, ok := insn.(openCall); ok {
			if inlinable[oc.name] {
				code = append(code[:i], append(nonterms[oc.name].Copy(), code[i+1:]...)...)
				continue
			}

			lbl, ok := labels[oc.name]
			if !ok {
				log.Fatal("Undefined non-terminal in grammar:", oc.name)
			}

			// replace this placeholder instruction with a normal call
			var replace isa.Insn = isa.Call{Lbl: lbl}
			// if a call is immediately followed by a return, optimize to
			// a jump for tail call optimization.
			next, ok := nextInsn(code[i+1:])
			if ok {
				switch next.(type) {
				case isa.Return:
					replace = isa.Jump{Lbl: lbl}
				}
			}

			// perform the replacement of the opencall by either a call or jump
			code[i] = replace
		}
	}

	code = append(code, LEnd)

	return code
}

// Copy creates a copy of this pattern so it can be used again
// within itself.
func (p Pattern) Copy() Pattern {
	// map from old labels to new labels
	code := make(Pattern, len(p))
	copy(code, p)
	labels := make(map[isa.Label]isa.Label)
	for i := range code {
		switch t := code[i].(type) {
		case isa.Label:
			l := isa.NewLabel()
			labels[t] = l
			code[i] = l
		}
	}
	for i := range code {
		switch t := code[i].(type) {
		case isa.Jump:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.Choice:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.Call:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.Commit:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.PartialCommit:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.BackCommit:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.TestChar:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.TestSet:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.TestAny:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.Choice2:
			t.Lbl = labels[t.Lbl]
			code[i] = t
		case isa.JumpType:
			panic("All jump types should be handled")
		}
	}
	return code
}

// Optimize performs some optimization passes on the code in p.
func (p Pattern) Optimize() {
	// map from label to index in code
	labels := make(map[isa.Label]int)
	for i, insn := range p {
		switch l := insn.(type) {
		case isa.Label:
			labels[l] = i
		}
	}

	for i, insn := range p {
		// head-fail optimization: if we find a choice instruction immediately
		// followed (no label) by Char/Set/Any, we can replace with the
		// dedicated instruction TestChar/TestSet/TestAny.
		if ch, ok := insn.(isa.Choice); ok && i < len(p)-1 {
			next := p[i+1]
			switch t := next.(type) {
			case isa.Char:
				p[i] = isa.TestChar{
					Byte: t.Byte,
					Lbl:  ch.Lbl,
				}
				p[i+1] = isa.Nop{}
			case isa.Set:
				p[i] = isa.TestSet{
					Chars: t.Chars,
					Lbl:   ch.Lbl,
				}
				p[i+1] = isa.Nop{}
			case isa.Any:
				p[i] = isa.TestAny{
					N:   t.N,
					Lbl: ch.Lbl,
				}
				p[i+1] = isa.Nop{}
			}
		}

		// jump optimization: if we find a jump to another control flow
		// instruction, we can replace the current jump directly with the
		// target instruction.
		if j, ok := insn.(isa.Jump); ok {
			next, ok := nextInsn(p[labels[j.Lbl]:])
			if ok {
				switch next.(type) {
				case isa.Call, isa.PartialCommit, isa.BackCommit, isa.Commit,
					isa.Jump, isa.Return, isa.Fail, isa.FailTwice, isa.End:
					p[i] = next
				}
			}
		}
	}
}

// String returns the string representation of the pattern.
func (p Pattern) String() string {
	s := ""
	for _, insn := range p {
		switch insn.(type) {
		case isa.Nop:
			continue
		case isa.Label:
			s += fmt.Sprintf("%v:", insn)
		default:
			s += fmt.Sprintf("\t%v\n", insn)
		}
	}
	s += "\n"
	return s
}

func nextInsn(p Pattern) (isa.Insn, bool) {
	for i := 0; i < len(p); i++ {
		switch p[i].(type) {
		case isa.Label:
			continue
		default:
			return p[i], true
		}
	}

	return isa.Nop{}, false
}
