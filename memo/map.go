package memo

// A MapTable implements a map-based memoization table
type MapTable map[Key]*Entry

// NewMapTable constructs a new MapTable.
func NewMapTable() MapTable {
	return make(MapTable)
}

// Get returns the memo entry associated with the given key.
func (t MapTable) Get(k Key) (*Entry, bool) {
	e, ok := t[k]
	return e, ok
}

// Put places a new memoization entry at the given key.
func (t MapTable) Put(k Key, e *Entry) {
	t[k] = e
}

// Delete deletes a memoization entry associated with a key.
func (t MapTable) Delete(k Key) {
	delete(t, k)
}

// ApplyEdit invalidates memoization entries that overlap with the edit, and
// updates memoization entries that are to the right of the edit.
func (t MapTable) ApplyEdit(e Edit) {
	size := e.Len - e.End.Cmp(e.Start)

	for key, ent := range t {
		x1, x2 := key.Pos, key.Pos.Move(ent.examined)
		y1, y2 := e.Start, e.End
		if x1.Cmp(y2) <= 0 && y1.Cmp(x2) <= 0 {
			delete(t, key)
			continue
		}

		// shift entries that are to the right of the edited interval
		if key.Pos.Cmp(e.Start) >= 0 {
			// Since the key is changed, we need to rehash the value
			delete(t, key)
			// TODO: not sure if this is fully correct
			key.Pos.Move(size)
			ent.start.Move(size)

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

// Size returns the number of memoization entries in the table.
func (t MapTable) Size() int {
	return len(t)
}
