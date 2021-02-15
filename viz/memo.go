package viz

import (
	"io"

	svg "github.com/ajstarks/svgo"
	"github.com/zyedidia/gpeg/memo"
)

func DrawMemo(table memo.Table, sz int, out io.Writer, width, height int) {
	loc := func(p int) int {
		return int(float64(p) / float64(sz) * float64(width))
	}

	entries := table.Overlaps(0, sz)

	canvas := svg.New(out)
	canvas.Start(width, height)

	h := 0
	for _, e := range entries {
		h += 3
		st := loc(e.Start())
		if e.Length() == -1 {
			canvas.Line(st, h, st+loc(e.Examined()), h, "stroke:magenta;stroke-width:2")
		} else {
			canvas.Line(st, h, st+loc(e.Length()), h, "stroke:blue;stroke-width:2")
			canvas.Line(st+loc(e.Length()), h, st+loc(e.Examined()), h, "stroke:red;stroke-width:2")
		}
	}

	canvas.End()
}
