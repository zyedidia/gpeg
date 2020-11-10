package pattern

import (
	"reflect"

	"github.com/zyedidia/gpeg/isa"
)

func Histogram(p isa.Program) map[reflect.Type]int {
	hist := make(map[reflect.Type]int)
	for _, insn := range p {
		t := reflect.TypeOf(insn)
		hist[t]++
	}
	return hist
}
