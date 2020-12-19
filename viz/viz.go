package viz

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"sort"

	"github.com/zyedidia/gpeg/isa"
)

// A Histogram maps instruction types to number of uses.
type Histogram map[reflect.Type]int

// String prints a histogram in sorted form.
func (h Histogram) String() string {
	type kv struct {
		Key   reflect.Type
		Value int
	}

	var ss []kv
	for k, v := range h {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	buf := &bytes.Buffer{}
	for _, kv := range ss {
		buf.WriteString(fmt.Sprintf("%v: %d\n", kv.Key, kv.Value))
	}
	return buf.String()
}

// ToHistogram returns the number of times each instruction occurs in the given
// parsing program.
func ToHistogram(p isa.Program) Histogram {
	hist := make(Histogram)
	for _, insn := range p {
		switch insn.(type) {
		case isa.Label, isa.Nop:
			continue
		}
		t := reflect.TypeOf(insn)
		hist[t]++
	}
	return hist
}

// WriteDotViz generates a PDF for the corresponding dot graph using the 'dot'
// program which must be installed. It writes the PDF content to filename.
func WriteDotViz(filename string, dot string) error {
	tmp, err := ioutil.TempFile("", "tmp.dot")
	if err != nil {
		return err
	}

	_, err = tmp.WriteString(dot)
	if err != nil {
		return err
	}

	out, err := exec.Command("dot", "-Tpdf", tmp.Name()).Output()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, out, 0644)
	if err != nil {
		return err
	}

	return os.Remove(tmp.Name())
}
