package gpeg

import (
	"io/ioutil"
	"testing"

	"github.com/zyedidia/gpeg/input/linerope"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

type point struct {
	line, col int
}

type edit struct {
	start point
	end   point
	text  string
}

// Open a 250k java file and apply some edits and verify that after each edit
// the incremental result is the same as doing a full parse.
func TestIncrementalJava(t *testing.T) {
	peg, err := ioutil.ReadFile("grammars/java.peg")
	if err != nil {
		t.Error(err)
	}
	p := re.MustCompile(string(peg))

	java, err := ioutil.ReadFile("testdata/test.java")
	if err != nil {
		t.Error(err)
	}

	edits := []edit{
		edit{point{229, 8}, point{229, 9}, ""},
		edit{point{229, 8}, point{229, 9}, ""},
		edit{point{229, 8}, point{229, 8}, "r"},
		edit{point{229, 8}, point{229, 8}, "e"},
		edit{point{902, 15}, point{902, 17}, "--"},
		edit{point{925, 10}, point{925, 10}, "t"},
		edit{point{925, 11}, point{925, 11}, "h"},
		edit{point{925, 12}, point{925, 12}, "e"},
		edit{point{925, 13}, point{925, 13}, " "},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 5}, ""},
		edit{point{1688, 4}, point{1688, 4}, "p"},
		edit{point{1688, 5}, point{1688, 5}, "u"},
		edit{point{1688, 6}, point{1688, 6}, "b"},
		edit{point{1688, 7}, point{1688, 7}, "l"},
		edit{point{1688, 8}, point{1688, 8}, "i"},
		edit{point{1688, 9}, point{1688, 9}, "c"},
	}

	tbl := memo.NewTreeTable(0)
	prog := pattern.MustCompile(p)
	code := vm.Encode(prog)

	r := linerope.New(java, &linerope.DefaultOptions)

	for i, e := range edits {
		start := r.OffsetAt(e.start.line-1, e.start.col)
		end := r.OffsetAt(e.end.line-1, e.end.col)

		r.Remove(start, end)
		r.Insert(start, []byte(e.text))

		tbl.ApplyEdit(memo.Edit{
			Start: start,
			End:   end,
			Len:   len(e.text),
		})

		match, off, _, _ := code.Exec(r, tbl)
		nmatch, noff, _, _ := code.Exec(r, memo.NoneTable{})

		if match != nmatch || off != noff {
			t.Fatal(i, match, nmatch, off, noff)
		}
	}
}
