package gpeg

type opCode byte

const (
	ipFail = -1
)

type vmstate struct {
	input   Reader
	ip      int // instruction pointer
	stack   *stack
	caplist []capture
}

func NewVM(r Reader) *vmstate {
	return &vmstate{
		input:   r,
		ip:      0,
		stack:   NewStack(),
		caplist: []capture{},
	}
}

func (vm *vmstate) exec(code []instr) int {
	for vm.ip < len(code) {
		if vm.ip == ipFail {
			ent, ok := vm.stack.Pop()
			if !ok {
				// match has failed
				return -1
			}
			if !ent.ReturnAddress() {
				vm.ip = ent.btrack.ip
				vm.caplist = ent.btrack.caplist
				vm.input.SeekBytes(ent.btrack.off, SeekStart)
			}
			// try again with new instruction pointer/stack value
			continue
		}
		code[vm.ip].exec(vm)
	}
	return vm.input.Offset()
}

type capture struct {
	subpos int
	inspos int
}
