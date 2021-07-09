package gpeg

import (
	"io/ioutil"
	"testing"

	"github.com/zyedidia/gpeg/bench"
	"github.com/zyedidia/gpeg/input/linerope"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

// Open a 250k java file and apply some edits and verify that after each edit
// the incremental result is the same as doing a full parse.
func TestIncrementalJava(t *testing.T) {
	peg, err := ioutil.ReadFile("grammars/java_memo.peg")
	if err != nil {
		t.Error(err)
	}
	p := re.MustCompile(string(peg))

	java, err := ioutil.ReadFile("testdata/test.java")
	if err != nil {
		t.Error(err)
	}

	edits := bench.GenerateEdits(java, 100)

	tbl := memo.NewTreeTable(0)
	prog := pattern.MustCompile(p)
	code := vm.Encode(prog)

	r := linerope.New(java, &linerope.DefaultOptions)

	for i, e := range edits {
		start := e.Start
		end := e.End

		r.Remove(start, end)
		r.Insert(start, []byte(e.Text))

		tbl.ApplyEdit(memo.Edit{
			Start: start,
			End:   end,
			Len:   len(e.Text),
		})

		match, off, _, _ := code.Exec(r, tbl)
		nmatch, noff, _, _ := code.Exec(r, memo.NoneTable{})

		if match != nmatch || off != noff {
			t.Fatal(i, match, nmatch, off, noff)
		}
	}
}
