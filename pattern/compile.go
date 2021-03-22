package pattern

import (
	"fmt"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
)

// A NotFoundError means a a non-terminal was not found during grammar
// compilation.
type NotFoundError struct {
	Name string
}

// Error returns the error message.
func (e *NotFoundError) Error() string { return "non-terminal " + e.Name + ": not found" }

// Compile takes an input pattern and returns the result of compiling it into a
// parsing program, and optimizing the program.
func Compile(p Pattern) (isa.Program, error) {
	c, err := p.Compile()
	if err != nil {
		return nil, err
	}

	Optimize(c)
	return c, nil
}

// MustCompile is the same as Compile but panics if there is an error during
// compilation.
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

// Compile this node.
func (p *AltNode) Compile() (isa.Program, error) {
	// optimization: if Left and Right are charsets/single chars, return the union
	set, ok := combine(Get(p.Left), Get(p.Right))
	if ok {
		return isa.Program{
			isa.Set{Chars: set},
		}, nil
	}

	l, err1 := Get(p.Left).Compile()
	r, err2 := Get(p.Right).Compile()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	L1 := isa.NewLabel()

	// optimization: if the right and left nodes are disjoint, we can use
	// NoChoice variants of the head-fail optimization instructions.
	var disjoint bool
	var testinsn isa.Insn
	linsn, okl := nextInsn(l)
	rinsn, okr := nextInsn(r)
	if okl && okr {
		switch lt := linsn.(type) {
		case isa.Set:
			switch rt := rinsn.(type) {
			case isa.Char:
				disjoint = !lt.Chars.Has(rt.Byte)
			}
			testinsn = isa.TestSetNoChoice{Chars: lt.Chars, Lbl: L1}
		case isa.Char:
			switch rt := rinsn.(type) {
			case isa.Char:
				disjoint = lt.Byte != rt.Byte
			case isa.Set:
				disjoint = !rt.Chars.Has(lt.Byte)
			}
			testinsn = isa.TestCharNoChoice{Byte: lt.Byte, Lbl: L1}
		}
	}

	L2 := isa.NewLabel()
	code := make(isa.Program, 0, len(l)+len(r)+5)
	if disjoint {
		code = append(code, testinsn)
		code = append(code, l[1:]...)
		code = append(code, isa.Jump{Lbl: L2})
	} else {
		code = append(code, isa.Choice{Lbl: L1})
		code = append(code, l...)
		code = append(code, isa.Commit{Lbl: L2})
	}
	code = append(code, L1)
	code = append(code, r...)
	code = append(code, L2)
	return code, nil
}

// Compile this node.
func (p *SeqNode) Compile() (isa.Program, error) {
	l, err1 := Get(p.Left).Compile()
	r, err2 := Get(p.Right).Compile()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	return append(l, r...), nil
}

// Compile this node.
func (p *StarNode) Compile() (isa.Program, error) {
	switch t := Get(p.Patt).(type) {
	case *ClassNode:
		// optimization: repeating a charset uses the dedicated instruction 'span'
		return isa.Program{
			isa.Span{Chars: t.Chars},
		}, nil
	case *MemoNode:
		// optimization: if the pattern we are repeating is a memoization
		// entry, we should use special instructions to memoize it as a tree to
		// get logarithmic saving when reparsing.
		sub, err := Get(t.Patt).Compile()
		code := make(isa.Program, 0, len(sub)+7)
		L1 := isa.NewLabel()
		L2 := isa.NewLabel()
		L3 := isa.NewLabel()
		NoJump := isa.NewLabel()

		code = append(code, L1)
		code = append(code, isa.MemoTreeOpen{Id: t.Id, Lbl: L3})
		code = append(code, isa.Choice{Lbl: L2})
		code = append(code, sub...)
		code = append(code, isa.Commit{Lbl: NoJump})
		code = append(code, NoJump)
		code = append(code, isa.MemoTreeInsert{})
		code = append(code, L3)
		code = append(code, isa.MemoTree{})
		code = append(code, isa.Jump{Lbl: L1})
		code = append(code, L2)
		code = append(code, isa.MemoTreeClose{Id: t.Id})

		// code = append(code, L1)
		// code = append(code, isa.MemoOpen{Id: t.Id, Lbl: L3})
		// code = append(code, isa.Choice{Lbl: L2})
		// code = append(code, sub...)
		// code = append(code, isa.Commit{Lbl: NoJump})
		// code = append(code, NoJump)
		// code = append(code, isa.MemoTree{})
		// code = append(code, L3)
		// code = append(code, isa.Jump{Lbl: L1})
		// code = append(code, L2)
		// code = append(code, isa.MemoTreeClose{})
		return code, err
	}

	sub, err := Get(p.Patt).Compile()
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

// Compile this node.
func (p *PlusNode) Compile() (isa.Program, error) {
	starp := Star(Get(p.Patt))
	star, err1 := starp.Compile()
	sub, err2 := Get(p.Patt).Compile()
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

// Compile this node.
func (p *OptionalNode) Compile() (isa.Program, error) {
	// optimization: if the pattern is a class node or single char literal, we
	// can use the Test*NoChoice instructions.
	switch t := Get(p.Patt).(type) {
	case *LiteralNode:
		if len(t.Str) == 1 {
			L1 := isa.NewLabel()
			return isa.Program{
				isa.TestCharNoChoice{Byte: t.Str[0], Lbl: L1},
				L1,
			}, nil
		}
	case *ClassNode:
		L1 := isa.NewLabel()
		prog := isa.Program{
			isa.TestSetNoChoice{Chars: t.Chars, Lbl: L1},
			L1,
		}
		return prog, nil
	}

	a := AltNode{
		Left:  Get(p.Patt),
		Right: &EmptyNode{},
	}
	return a.Compile()
}

// Compile this node.
func (p *NotNode) Compile() (isa.Program, error) {
	sub, err := Get(p.Patt).Compile()
	L1 := isa.NewLabel()
	code := make(isa.Program, 0, len(sub)+3)
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, sub...)
	code = append(code, isa.FailTwice{})
	code = append(code, L1)
	return code, err
}

// Compile this node.
func (p *AndNode) Compile() (isa.Program, error) {
	sub, err := Get(p.Patt).Compile()
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

// Compile this node.
func (p *CapNode) Compile() (isa.Program, error) {
	sub, err := Get(p.Patt).Compile()
	if err != nil {
		return nil, err
	}
	code := make(isa.Program, 0, len(sub)+2)

	i := 0
	back := 0
loop:
	for _, insn := range sub {
		switch t := insn.(type) {
		case isa.Char, isa.Set:
			back++
		case isa.Any:
			back += int(t.N)
		default:
			break loop
		}
		i++
	}

	if i == 0 || back >= 256 {
		code = append(code, isa.CaptureBegin{Id: p.Id})
		i = 0
	} else if i == len(sub) && back < 256 {
		code = append(code, sub...)
		code = append(code, isa.CaptureFull{Back: byte(back), Id: p.Id})
		return code, nil
	} else {
		code = append(code, sub[:i]...)
		code = append(code, isa.CaptureLate{Back: byte(back), Id: p.Id})
	}
	code = append(code, sub[i:]...)
	code = append(code, isa.CaptureEnd{})
	return code, nil
}

// Compile this node.
func (p *MemoNode) Compile() (isa.Program, error) {
	L1 := isa.NewLabel()
	sub, err := Get(p.Patt).Compile()
	code := make(isa.Program, 0, len(sub)+3)
	code = append(code, isa.MemoOpen{Lbl: L1, Id: p.Id})
	code = append(code, sub...)
	code = append(code, isa.MemoClose{})
	code = append(code, L1)
	return code, err
}

// Compile this node.
func (p *CheckNode) Compile() (isa.Program, error) {
	L1 := isa.NewLabel()
	sub, err := Get(p.Patt).Compile()
	code := make(isa.Program, 0, len(sub)+3)
	code = append(code, isa.CheckBegin{})
	code = append(code, sub...)
	code = append(code, isa.CheckEnd{Checker: p.Checker})
	code = append(code, L1)
	return code, err
}

// Compile this node.
func (p *SearchNode) Compile() (isa.Program, error) {
	var rsearch Pattern
	var set charset.Set
	opt := false

	sub, err := Get(p.Patt).Compile()
	if err != nil {
		return nil, err
	}

	next, ok := nextInsn(sub)
	if ok {
		switch t := next.(type) {
		case isa.Char:
			set = charset.New([]byte{t.Byte}).Complement()
			opt = true
		case isa.Set:
			// Heuristic: if the set is smaller than 10 chars, it
			// is unlikely enough to match that we should consume all
			// chars from the complement before continuing the search.
			// The number 10 was arbitrarily chosen.
			if t.Chars.Size() < 10 {
				set = t.Chars.Complement()
				opt = true
			}
		}
	}

	if opt {
		rsearch = Concat(Star(Set(set)), NonTerm("S"))
	} else {
		rsearch = NonTerm("S")
	}

	return Grammar("S", map[string]Pattern{
		"S": Or(Get(p.Patt), Concat(Any(1), rsearch)),
	}).Compile()
}

// Compile this node.
func (p *GrammarNode) Compile() (isa.Program, error) {
	p.Inline()

	used := make(map[string]bool)
	for _, v := range p.Defs {
		WalkPattern(v, true, func(sub Pattern) {
			switch t := sub.(type) {
			case *NonTermNode:
				if t.Inlined == nil {
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

// Compile this node.
func (p *ClassNode) Compile() (isa.Program, error) {
	return isa.Program{
		isa.Set{Chars: p.Chars},
	}, nil
}

// Compile this node.
func (p *LiteralNode) Compile() (isa.Program, error) {
	code := make(isa.Program, len(p.Str))
	for i := 0; i < len(p.Str); i++ {
		code[i] = isa.Char{Byte: p.Str[i]}
	}
	return code, nil
}

// Compile this node.
func (p *NonTermNode) Compile() (isa.Program, error) {
	if p.Inlined != nil {
		return p.Inlined.Compile()
	}
	return isa.Program{
		openCall{name: p.Name},
	}, nil
}

// Compile this node.
func (p *DotNode) Compile() (isa.Program, error) {
	return isa.Program{
		isa.Any{N: p.N},
	}, nil
}

// Compile this node.
func (p *ErrorNode) Compile() (isa.Program, error) {
	var recovery isa.Program
	var err error

	if p.Recover == nil {
		recovery = isa.Program{
			isa.End{Fail: true},
		}
	} else {
		recovery, err = Get(p.Recover).Compile()
	}

	code := make(isa.Program, 0, len(recovery)+1)
	code = append(code, isa.Error{Message: p.Message})
	code = append(code, recovery...)
	return code, err
}

// Compile this node.
func (p *EmptyNode) Compile() (isa.Program, error) {
	return isa.Program{}, nil
}
