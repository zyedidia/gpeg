package vm

import (
	"github.com/zyedidia/gpeg/input"
)

type capt struct {
	off input.Pos
	ip  int
}

func (vm *VM) Captures(capt []capt, code VMCode) [][]byte {
	ind := vm.CapturesIndex(capt, code)
	caps := make([][]byte, len(ind))
	for i, c := range ind {
		caps[i] = vm.input.Slice(c[0], c[1])
	}
	return caps
}

func (vm *VM) CapturesString(capt []capt, code VMCode) []string {
	ind := vm.CapturesIndex(capt, code)
	caps := make([]string, len(ind))
	for i, c := range ind {
		caps[i] = string(vm.input.Slice(c[0], c[1]))
	}
	return caps
}

func (vm *VM) CapturesIndex(capt []capt, code VMCode) [][2]input.Pos {
	stack := newCapStack()
	caps := make([][2]input.Pos, 0, len(capt))
	for _, c := range capt {
		op := code.insns[c.ip]

		switch op {
		case opCaptureBegin, opCaptureLate:
			stack.push(c)
		case opCaptureEnd:
			ent, ok := stack.pop()
			if !ok {
				panic("Error: capture closed but not opened")
			}
			caps = append(caps, [2]input.Pos{ent.off, c.off})
		case opCaptureFull:
			back := decodeByte(code.insns[c.ip+1:])
			caps = append(caps, [2]input.Pos{c.off, c.off + input.Pos(back)})
		}
	}
	return caps
}

type capstack struct {
	entries []capt
}

func newCapStack() *capstack {
	return &capstack{
		entries: make([]capt, 0, 4),
	}
}

func (s *capstack) reset() {
	s.entries = s.entries[:1]
}

func (s *capstack) push(ent capt) {
	s.entries = append(s.entries, ent)
}

func (s *capstack) pop() (*capt, bool) {
	if len(s.entries) == 0 {
		return nil, false
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return &ret, true
}

func (s *capstack) peek() *capt {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}
