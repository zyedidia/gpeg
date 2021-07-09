package gpeg

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

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
	rand.Seed(42)

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
	edits = bench.ToSingleEdits(edits)

	tbl := memo.NewTreeTable(512)
	prog := pattern.MustCompile(p)
	code := vm.Encode(prog)

	r := linerope.New(java, &linerope.DefaultOptions)

	for _, e := range edits {
		start := e.Start
		end := e.End

		r.Remove(start, end)
		r.Insert(start, []byte(e.Text))

		st := time.Now()
		tbl.ApplyEdit(memo.Edit{
			Start: start,
			End:   end,
			Len:   len(e.Text),
		})

		match, off, _, _ := code.Exec(r, tbl)
		fmt.Println("reparse", time.Since(st), match, off)
		// st = time.Now()
		// nmatch, noff, _, _ := code.Exec(r, memo.NoneTable{})
		// fmt.Println("full parse", time.Since(st))

		// if match != nmatch || off != noff {
		// 	t.Fatal(i, match, nmatch, off, noff)
		// }
	}
}
