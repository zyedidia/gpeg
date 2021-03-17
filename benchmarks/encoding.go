package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

var grammar = flag.String("grammar", "", "grammar to test")
var ins = flag.Bool("ins", false, "show instruction data size")

func main() {
	flag.Parse()

	peg, err := ioutil.ReadFile(*grammar)
	if err != nil {
		log.Fatal(err)
	}

	inline := []int{0, 50, 100, 150, 200, 250, 300, 350, 400, 450}

	for _, t := range inline {
		pattern.InlineThreshold = t

		p := re.MustCompilePatt(string(peg))
		prog := pattern.MustCompile(p)
		code := vm.Encode(prog)

		if *ins {
			fmt.Println(t, code.Size())
		} else {
			b, err := code.ToBytes()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(t, len(b))
		}
	}
}
