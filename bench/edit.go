package bench

import (
	"math/rand"

	"github.com/zyedidia/gpeg/input/linerope"
	"github.com/zyedidia/gpeg/memo"
	p "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

type Edit struct {
	Start, End int
	Text       []byte
}

func EditToEdits(e Edit) []Edit {
	var edits []Edit

	for i := e.Start; i < e.End; i++ {
		edits = append(edits, Edit{
			Start: e.Start,
			End:   e.Start + 1,
			Text:  nil,
		})
	}

	for i := 0; i < len(e.Text); i++ {
		edits = append(edits, Edit{
			Start: e.Start + i,
			End:   e.Start + i,
			Text:  []byte{e.Text[i]},
		})
	}

	return edits
}

func ToSingleEdits(edits []Edit) []Edit {
	single := make([]Edit, 0)

	for _, e := range edits {
		single = append(single, EditToEdits(e)...)
	}

	return single
}

// strategies for generating edits to a Java file:
// * insert newline at start of line
// * change contents of comment
// * delete single-line comment
// * change function name
// * change function qualifier (e.g., from 'private' to 'public')
// * change contents of string

type EditType int

const (
	EditInsertNewline EditType = iota
	EditRemoveNewline
	EditWhitespace
	EditChangeComment
	EditRemoveComment
	EditChangeFunc
	EditChangeFuncQual
	EditChangeString
)

var editTypes = []EditType{
	EditInsertNewline,
	EditRemoveNewline,
	EditChangeComment,
	EditRemoveComment,
	EditChangeFunc,
	EditChangeFuncQual,
	EditChangeString,
}

func GenerateEdits(data []byte, nedits int) []Edit {
	r := linerope.New(data)
	edits := make([]Edit, 0, nedits)

	prog := p.MustCompile(grammar)
	java := vm.Encode(prog)
	tbl := memo.NewTreeTable(512)

	for i := 0; i < nedits; {
		_, _, ast, _ := java.Exec(r, tbl)

		var e Edit
		typ := editTypes[rand.Intn(len(editTypes))]

		switch typ {
		case EditInsertNewline:
			line := rand.Intn(r.NumLines())
			off := r.OffsetAt(line, 0)
			e = Edit{
				Start: off,
				End:   off,
				Text:  []byte{'\n'},
			}
		case EditRemoveNewline:
			candidates := make([]*memo.Capture, 0)
			it := ast.ChildIterator(0)
			for ch := it(); ch != nil; ch = it() {
				if ch.Id() == capNewline {
					candidates = append(candidates, ch)
				}
			}
			if len(candidates) == 0 {
				continue
			}
			ch := candidates[rand.Intn(len(candidates))]
			e = Edit{
				Start: ch.Start(),
				End:   ch.Start() + ch.Len(),
				Text:  nil,
			}
		case EditRemoveComment:
			candidates := make([]*memo.Capture, 0)
			it := ast.ChildIterator(0)
			for ch := it(); ch != nil; ch = it() {
				if ch.Id() == capLineComment {
					candidates = append(candidates, ch)
				}
			}
			if len(candidates) == 0 {
				continue
			}
			ch := candidates[rand.Intn(len(candidates))]
			line, _ := r.LineColAt(ch.Start())
			e = Edit{
				Start: ch.Start(),
				End:   r.OffsetAt(line+1, 0),
				Text:  nil,
			}
		case EditChangeFunc:
			candidates := make([]*memo.Capture, 0)
			it := ast.ChildIterator(0)
			for ch := it(); ch != nil; ch = it() {
				if ch.Id() == capFuncName {
					candidates = append(candidates, ch)
				}
			}
			if len(candidates) == 0 {
				continue
			}
			ch := candidates[rand.Intn(len(candidates))]
			e = Edit{
				Start: ch.Start(),
				End:   ch.Start() + ch.Len(),
				Text:  randID(rand.Intn(5) + 4),
			}
		case EditChangeFuncQual:
			candidates := make([]*memo.Capture, 0)
			it := ast.ChildIterator(0)
			for ch := it(); ch != nil; ch = it() {
				if ch.Id() == capFuncQual {
					candidates = append(candidates, ch)
				}
			}
			if len(candidates) == 0 {
				continue
			}
			ch := candidates[rand.Intn(len(candidates))]
			e = Edit{
				Start: ch.Start(),
				End:   ch.Start() + ch.Len(),
				Text:  []byte("protected"),
			}
		default:
			continue
		}

		r.Remove(e.Start, e.End)
		r.Insert(e.Start, e.Text)
		tbl.ApplyEdit(memo.Edit{
			Start: e.Start,
			End:   e.End,
			Len:   len(e.Text),
		})

		edits = append(edits, e)
		i++
	}

	// r.WriteTo(os.Stdout)

	return edits
}

var rbytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randID(n int) []byte {
	id := make([]byte, n)
	for i := range id {
		id[i] = rbytes[rand.Intn(len(rbytes))]
	}
	return id
}
