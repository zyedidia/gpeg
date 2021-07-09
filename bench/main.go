// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/zyedidia/gpeg/bench"
)

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	edits := bench.GenerateEdits(data, 100)

	for _, e := range edits {
		fmt.Printf("(%d, %d): %s\n", e.Start, e.End, strconv.Quote(string(e.Text)))
	}
}
