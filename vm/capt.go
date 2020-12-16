package vm

import (
	"github.com/zyedidia/gpeg/ast"

	"github.com/zyedidia/gpeg/input"
)

type capt struct {
	off input.Pos
	ip  int
}

func (vm *VM) Captures(capt []*ast.Node) [][]byte {
	ind := vm.CapturesIndex(capt)
	caps := make([][]byte, len(ind))
	for i, c := range ind {
		caps[i] = vm.input.Slice(c[0], c[1])
	}
	return caps
}

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
