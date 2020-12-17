package memo

import (
	"github.com/zyedidia/gpeg/ast"
	"github.com/zyedidia/gpeg/input"
)

type MapTable map[Key]Entry

func NewMapTable() MapTable {
	return make(MapTable)
}

func (t MapTable) Get(k Key) (Entry, bool) {
	e, ok := t[k]
	return e, ok
}

func (t MapTable) Put(k Key, e Entry) {
	t[k] = e
}

func (t MapTable) Delete(k Key) {
	delete(t, k)
}

func (t MapTable) ApplyEdit(e Edit) {
	size := e.Len - int(e.End-e.Start)

	for key, ent := range t {
		x1, x2 := int(key.Pos), int(key.Pos)+ent.examined
		y1, y2 := int(e.Start), int(e.End)
		if x1 <= y2 && y1 <= x2 {
			delete(t, key)
			continue
		}

		// shift entries that are to the right of the edited interval
		if key.Pos >= e.Start {
			// Since the key is changed, we need to rehash the value
			delete(t, key)
			// TODO: not sure if this is fully correct
			key.Pos += input.Pos(size)

			for _, n := range ent.Value() {
				n.Each(func(child *ast.Node) {
					child.Advance(size)
				})
			}

			// If there is an entry where we want to move to, that is
			// unfortunate we will just invalidate this entry, even though
			// it's not strictly necessary for simplicity (hopefully this
			// does not happen often)
			_, ok := t[key]
			if !ok {
				t[key] = ent
			}
		}
	}
}

func (t MapTable) Size() int {
	return len(t)
}
