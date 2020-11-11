package pattern

import (
	"fmt"

	"github.com/zyedidia/gpeg/isa"
)

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string { return "non-terminal " + e.Name + ": not found" }

func Compile(p Pattern) (isa.Program, error) {
	c, err := p.Compile()
	if err != nil {
		return nil, err
	}

	Optimize(c)
	return c, nil
}

func MustCompile(p Pattern) isa.Program {
	c, err := Compile(p)
	if err != nil {
		panic(err)
	}
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

func (p *AltNode) Compile() (isa.Program, error) {
	// optimization: if p1 and p2 are charsets, return the union
	cl, okl := p.Left.(*ClassNode)
	cr, okr := p.Right.(*ClassNode)
	if okl && okr {
		return isa.Program{
			isa.Set{Chars: cl.Chars.Add(cr.Chars)},
		}, nil
	}

	l, err1 := p.Left.Compile()
	r, err2 := p.Right.Compile()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	code := make(isa.Program, 0, len(l)+len(r)+5)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, l...)
	code = append(code, isa.Commit{Lbl: L2})
	code = append(code, L1)
	code = append(code, r...)
	code = append(code, L2)
	return code, nil
}

func (p *SeqNode) Compile() (isa.Program, error) {
	l, err1 := p.Left.Compile()
	r, err2 := p.Right.Compile()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	return append(l, r...), nil
}

func (p *StarNode) Compile() (isa.Program, error) {
	// optimization: repeating a charset uses the dedicated instruction 'span'
	switch t := p.Patt.(type) {
	case *ClassNode:
		return isa.Program{
			isa.Span{Chars: t.Chars},
		}, nil
	}

	sub, err := p.Patt.Compile()
	code := make(isa.Program, 0, len(sub)+4)

	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, L1)
	code = append(code, sub...)
	code = append(code, isa.PartialCommit{Lbl: L1})
	code = append(code, L2)
	return code, err
}

func (p *PlusNode) Compile() (isa.Program, error) {
	starp := StarNode{
		Patt: p.Patt,
	}
	star, err1 := starp.Compile()
	sub, err2 := p.Patt.Compile()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	code := make(isa.Program, 0, len(sub)+len(star))
	code = append(code, sub...)
	code = append(code, star...)
	return code, nil
}

func (p *OptionalNode) Compile() (isa.Program, error) {
	a := AltNode{
		Left:  p.Patt,
		Right: &EmptyNode{},
	}
	return a.Compile()
}

func (p *NotNode) Compile() (isa.Program, error) {
	sub, err := p.Patt.Compile()
	L1 := isa.NewLabel()
	code := make(isa.Program, 0, len(sub)+3)
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, sub...)
	code = append(code, isa.FailTwice{})
	code = append(code, L1)
	return code, err
}

func (p *AndNode) Compile() (isa.Program, error) {
	sub, err := p.Patt.Compile()
	code := make(isa.Program, 0, len(sub)+5)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()

	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, sub...)
	code = append(code, isa.BackCommit{Lbl: L2})
	code = append(code, L1)
	code = append(code, isa.Fail{})
	code = append(code, L2)
	return code, err
}

func (p *CapNode) Compile() (isa.Program, error) {
	sub, err := p.Patt.Compile()
	code := make(isa.Program, 0, len(sub)+2)
	code = append(code, isa.CaptureBegin{Id: p.Id})
	code = append(code, sub...)
	code = append(code, isa.CaptureEnd{})
	return code, err
}

func (p *MemoNode) Compile() (isa.Program, error) {
	L1 := isa.NewLabel()
	sub, err := p.Patt.Compile()
	code := make(isa.Program, 0, len(sub)+3)
	code = append(code, isa.MemoOpen{Lbl: L1, Id: p.Id})
	code = append(code, sub...)
	code = append(code, isa.MemoClose{})
	code = append(code, L1)
	return code, err
}

func (p *GrammarNode) Compile() (isa.Program, error) {
	p.Inline()

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

	code := make(isa.Program, 0)
	LEnd := isa.NewLabel()
	code = append(code, openCall{name: p.Start}, isa.Jump{Lbl: LEnd})

	labels := make(map[string]isa.Label)
	for k, v := range p.Defs {
		if k != p.Start && !used[k] {
			continue
		}
		label := isa.NewLabel()
		labels[k] = label
		fn, err := v.Compile()
		if err != nil {
			return nil, err
		}
		code = append(code, label)
		code = append(code, fn...)
		code = append(code, isa.Return{})
	}

	// resolve calls to openCall and do tail call optimization
	for i := 0; i < len(code); i++ {
		insn := code[i]
		if oc, ok := insn.(openCall); ok {
			lbl, ok := labels[oc.name]
			if !ok {
				return nil, &NotFoundError{
					Name: oc.name,
				}
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

	return code, nil
}

func (p *ClassNode) Compile() (isa.Program, error) {
	return isa.Program{
		isa.Set{Chars: p.Chars},
	}, nil
}

func (p *LiteralNode) Compile() (isa.Program, error) {
	code := make(isa.Program, len(p.Str))
	for i := 0; i < len(p.Str); i++ {
		code[i] = isa.Char{Byte: p.Str[i]}
	}
	return code, nil
}

func (p *NonTermNode) Compile() (isa.Program, error) {
	if p.inlined != nil {
		return p.inlined.Compile()
	}
	return isa.Program{
		openCall{name: p.Name},
	}, nil
}

func (p *DotNode) Compile() (isa.Program, error) {
	return isa.Program{
		isa.Any{N: p.N},
	}, nil
}

func (p *EmptyNode) Compile() (isa.Program, error) {
	code := make(isa.Program, 0)
	return code, nil
}
