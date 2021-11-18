package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
)

func main() {
	flag.Parse()

	args := flag.Args()

	var in io.Reader
	if len(args) <= 0 {
		in = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		in = f
	}

	bytes, err := io.ReadAll(in)
	if err != nil {
		log.Fatal(err)
	}
	patt, err := re.Compile(string(bytes))
	if err != nil {
		log.Fatal(err)
	}
	prog, err := pattern.Compile(patt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(prog)
}
