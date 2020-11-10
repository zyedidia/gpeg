package pattern

import (
	"fmt"
	"log"

	"github.com/zyedidia/gpeg/isa"
)

type Compiled []isa.Insn

func Compile(p Pattern) Compiled {
	c := p.Compile()
	c.Optimize()
	return c
}

// openCall is a dummy instruction for resolving recursive function calls in
// grammars.
type openCall struct {
	name string
	isa.Nop
}

func (i openCall) String() string {
	return fmt.Sprintf("OpenCall %v", i.name)
}

func (p *AltNode) Compile() Compiled {
	// optimization: if p1 and p2 are charsets, return the union
	cl, okl := p.Left.(*ClassNode)
	cr, okr := p.Right.(*ClassNode)
	if okl && okr {
		return Compiled{
			isa.Set{Chars: cl.Chars.Add(cr.Chars)},
		}
	}

	l, r := p.Left.Compile(), p.Right.Compile()
	code := make(Compiled, 0, len(l)+len(r)+5)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, l...)
	code = append(code, isa.Commit{Lbl: L2})
	code = append(code, L1)
	code = append(code, r...)
	code = append(code, L2)
	return code
}

func (p *SeqNode) Compile() Compiled {
	l, r := p.Left.Compile(), p.Right.Compile()
	return append(l, r...)
}

func (p *StarNode) Compile() Compiled {
	// optimization: repeating a charset uses the dedicated instruction 'span'
	switch t := p.Patt.(type) {
	case *ClassNode:
		return Compiled{
			isa.Span{Chars: t.Chars},
		}
	}

	sub := p.Patt.Compile()
	code := make(Compiled, 0, len(sub)+4)

	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, L1)
	code = append(code, sub...)
	code = append(code, isa.PartialCommit{Lbl: L1})
	code = append(code, L2)
	return code
}

func (p *PlusNode) Compile() Compiled {
	starp := StarNode{
		Patt: p.Patt,
	}
	star := starp.Compile()
	sub := p.Patt.Compile()
	code := make(Compiled, 0, len(sub)+len(star))
	code = append(code, sub...)
	code = append(code, star...)
	return code
}

func (p *OptionalNode) Compile() Compiled {
	a := AltNode{
		Left:  p.Patt,
		Right: &EmptyNode{},
	}
	return a.Compile()
}

func (p *NotNode) Compile() Compiled {
	sub := p.Patt.Compile()
	L1 := isa.NewLabel()
	code := make(Compiled, 0, len(sub)+3)
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, sub...)
	code = append(code, isa.FailTwice{})
	code = append(code, L1)
	return code
}

func (p *AndNode) Compile() Compiled {
	sub := p.Patt.Compile()
	code := make(Compiled, 0, len(sub)+5)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()

	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, sub...)
	code = append(code, isa.BackCommit{Lbl: L2})
	code = append(code, L1)
	code = append(code, isa.Fail{})
	code = append(code, L2)
	return code
}

func (p *CapNode) Compile() Compiled {
	sub := p.Patt.Compile()
	code := make(Compiled, 0, len(sub)+2)
	code = append(code, isa.CaptureBegin{Id: p.Id})
	code = append(code, sub...)
	code = append(code, isa.CaptureEnd{})
	return code
}

func (p *MemoNode) Compile() Compiled {
	L1 := isa.NewLabel()
	sub := p.Patt.Compile()
	code := make(Compiled, 0, len(sub)+3)
	code = append(code, isa.MemoOpen{Lbl: L1, Id: p.Id})
	code = append(code, sub...)
	code = append(code, isa.MemoClose{})
	code = append(code, L1)
	return code
}

func (p *GrammarNode) Compile() Compiled {
	for p.inline() {
	}

	used := make(map[string]bool)
	for _, v := range p.Defs {
		WalkPattern(v, func(sub Pattern) {
			switch t := sub.(type) {
			case *NonTermNode:
				if t.inlined == nil {
					used[t.Name] = true
				}
			}
		})
	}

	if len(used) == 0 {
		return p.Defs[p.Start].Compile()
	}

	code := make(Compiled, 0)
	LEnd := isa.NewLabel()
	code = append(code, openCall{name: p.Start}, isa.Jump{Lbl: LEnd})

	labels := make(map[string]isa.Label)
	for k, v := range p.Defs {
		if k != p.Start && !used[k] {
			continue
		}
		label := isa.NewLabel()
		labels[k] = label
		code = append(code, label)
		code = append(code, v.Compile()...)
		code = append(code, isa.Return{})
	}

	// resolve calls to openCall and do tail call optimization
	for i := 0; i < len(code); i++ {
		insn := code[i]
		if oc, ok := insn.(openCall); ok {
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
					// remove the return instruction if there is no label referring to it
					retidx, hadlbl := nextInsnLabel(code[i+1:])
					if !hadlbl {
						code[i+1+retidx] = isa.Nop{}
					}
				}
			}

			// perform the replacement of the opencall by either a call or jump
			code[i] = replace
		}
	}

	code = append(code, LEnd)

	return code
}

func (p *ClassNode) Compile() Compiled {
	return Compiled{
		isa.Set{Chars: p.Chars},
	}
}

func (p *LiteralNode) Compile() Compiled {
	code := make(Compiled, len(p.Str))
	for i := 0; i < len(p.Str); i++ {
		code[i] = isa.Char{Byte: p.Str[i]}
	}
	return code
}

func (p *NonTermNode) Compile() Compiled {
	if p.inlined != nil {
		return p.inlined.Compile()
	}
	return Compiled{
		openCall{name: p.Name},
	}
}

func (p *DotNode) Compile() Compiled {
	return Compiled{
		isa.Any{N: 1},
	}
}

func (p *EmptyNode) Compile() Compiled {
	code := make(Compiled, 0)
	return code
}

// String returns the string representation of the pattern.
func (p Compiled) String() string {
	s := ""
	var last isa.Insn
	for _, insn := range p {
		switch insn.(type) {
		case isa.Nop:
			continue
		case isa.Label:
			if _, ok := last.(isa.Label); ok {
				s += "\rL...:"
			} else {
				s += fmt.Sprintf("%v:", insn)
			}
		default:
			s += fmt.Sprintf("\t%v\n", insn)
		}
		last = insn
	}
	s += "\n"
	return s
}
