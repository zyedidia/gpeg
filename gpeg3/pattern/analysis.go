package pattern

import "reflect"

func Histogram(p Compiled) map[reflect.Type]int {
	hist := make(map[reflect.Type]int)
	for _, insn := range p {
		t := reflect.TypeOf(insn)
		hist[t]++
	}
	return hist
}
