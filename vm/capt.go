package vm

import (
	"github.com/zyedidia/gpeg/ast"

	"github.com/zyedidia/gpeg/input"
)

type capt struct {
	off input.Pos
	ip  int
}

// func (vm *VM) Captures(capt []capt, code VMCode) [][]byte {
// 	ind := vm.CapturesIndex(capt, code)
// 	caps := make([][]byte, len(ind))
// 	for i, c := range ind {
// 		caps[i] = vm.input.Slice(c[0], c[1])
// 	}
// 	return caps
// }

func (vm *VM) CapturesString(capt []*ast.Node) []string {
	ind := vm.CapturesIndex(capt)
	caps := make([]string, len(ind))
	for i, c := range ind {
		caps[i] = string(vm.input.Slice(c[0], c[1]))
	}
	return caps
}

func (vm *VM) CapturesIndex(capt []*ast.Node) [][2]input.Pos {
	caps := make([][2]input.Pos, 0, len(capt))
	for _, c := range capt {
		caps = append(caps, [2]input.Pos{
			c.Start, c.End,
		})
		if c.Children != nil {
			caps = append(caps, vm.CapturesIndex(c.Children)...)
		}
	}
	return caps
}

func (vm *VM) CaptureAST(capt []capt, code VMCode) []*ast.Node {
	stack := newCapStack()
	nodes := make([]*ast.Node, 0)
	for _, c := range capt {
		op := code.data.Insns[c.ip]

		switch op {
		case opCaptureBegin, opCaptureLate:
			id := decodeI16(code.data.Insns[c.ip+2:])
			stack.push(astCapt{
				off: c.off,
				ip:  c.ip,
				node: &ast.Node{
					Id:       id,
					Start:    c.off,
					End:      0,
					Children: make([]*ast.Node, 0),
				},
			})
		case opCaptureEnd:
			ent, ok := stack.pop()
			if !ok {
				panic("Error: capture closed but not opened")
			}
			ent.node.End = c.off
			prev := stack.peek()
			if prev == nil {
				nodes = append(nodes, ent.node)
			} else {
				prev.node.Children = append(prev.node.Children, ent.node)
			}
			// case opCaptureFull:
			// 	back := decodeU8(code.data.Insns[c.ip+1:])
			// 	caps = append(caps, [2]input.Pos{c.off, c.off - input.Pos(back)})
		}
	}
	return nodes
}

type astCapt struct {
	off  input.Pos
	ip   int
	node *ast.Node
}

type capstack struct {
	entries []astCapt
}

func newCapStack() *capstack {
	return &capstack{
		entries: make([]astCapt, 0, 4),
	}
}

func (s *capstack) reset() {
	s.entries = s.entries[:1]
}

func (s *capstack) push(ent astCapt) {
	s.entries = append(s.entries, ent)
}

func (s *capstack) pop() (*astCapt, bool) {
	if len(s.entries) == 0 {
		return nil, false
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return &ret, true
}

func (s *capstack) peek() *astCapt {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}
