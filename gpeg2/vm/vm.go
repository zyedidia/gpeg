package vm

type VM struct {
	ip    int
	st    stack
	input input.BufReader
}

func NewVM(r input.BufReader) *VM {
	return &VM{
		ip:    0,
		st:    newStack(),
		input: r,
	}
}

func (v *VM) Exec(code VMCode) int {

}
