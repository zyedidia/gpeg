package viz

import (
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"

	"github.com/zyedidia/gpeg/isa"
)

// Histogram returns the number of times each instruction occurs in the given
// parsing program.
func Histogram(p isa.Program) map[reflect.Type]int {
	hist := make(map[reflect.Type]int)
	for _, insn := range p {
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
