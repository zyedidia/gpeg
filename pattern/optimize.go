package pattern

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

const InlineThreshold = 100

// Inline performs inlining passes until the inliner reaches a steady-state.
func (p *GrammarNode) Inline() {
	for p.inline() {
	}
}

func (p *GrammarNode) inline() bool {
	sizes := make(map[string]int)
	leaves := make(map[string]bool)
	for n, sub := range p.Defs {
		size := 0
		leaf := true
		WalkPattern(sub, true, func(s Pattern) {
			switch t := s.(type) {
			case *NonTermNode:
				if t.inlined == nil {
					leaf = false
				}
			}
			size++
		})
		sizes[n] = size
		leaves[n] = leaf
	}

	didInline := false
	WalkPattern(p, true, func(sub Pattern) {
		switch t := sub.(type) {
		case *NonTermNode:
			if sz, ok := sizes[t.Name]; ok && t.inlined == nil {
				if sz < InlineThreshold && leaves[t.Name] {
					didInline = true
					t.inlined = p.Defs[t.Name]
				}
			}
		}
	})
	return didInline
}

// If the bytes matched by p1 and p2 can be matched by a single charset, then
// that single combined charset is returned.
func combine(p1 Pattern, p2 Pattern) (charset.Set, bool) {
	var set charset.Set
	switch t1 := p1.(type) {
	case *LiteralNode:
		if len(t1.Str) != 1 {
			return set, false
		}
		switch t2 := p2.(type) {
		case *ClassNode:
			return t2.Chars.Add(charset.New([]byte{t1.Str[0]})), true
		case *LiteralNode:
			if len(t2.Str) != 1 {
				return set, false
			}
			return charset.New([]byte{t1.Str[0], t2.Str[0]}), true
		}
	case *ClassNode:
		switch t2 := p2.(type) {
		case *ClassNode:
			return t2.Chars.Add(t1.Chars), true
		case *LiteralNode:
			if len(t2.Str) != 1 {
				return set, false
			}
			return t1.Chars.Add(charset.New([]byte{t2.Str[0]})), true
		}
	}
	return set, false
}

func nextInsn(p isa.Program) (isa.Insn, bool) {
	for i := 0; i < len(p); i++ {
		switch p[i].(type) {
		case isa.Label, isa.Nop:
			continue
		default:
			return p[i], true
		}
	}

	return isa.Nop{}, false
}

func nextInsnLabel(p isa.Program) (int, bool) {
	hadLabel := false
	for i := 0; i < len(p); i++ {
		switch p[i].(type) {
		case isa.Nop:
			continue
		case isa.Label:
			hadLabel = true
		default:
			return i, hadLabel
		}
	}

	return -1, hadLabel
}

// Optimize performs some optimization passes on the code in p.
func Optimize(p isa.Program) {
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
				case isa.PartialCommit, isa.BackCommit, isa.Commit,
					isa.Jump, isa.Return, isa.Fail, isa.FailTwice, isa.End:
					p[i] = next
				}
			}
		}
	}
}