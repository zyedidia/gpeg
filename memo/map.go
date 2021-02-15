package memo

import (
	"sort"
	"sync"
)

type key struct {
	id  int
	pos int
}

var lock sync.Mutex

// A MapTable implements a map-based memoization table
type MapTable map[key][]*Entry

// NewMapTable constructs a new MapTable.
func NewMapTable() MapTable {
	return make(MapTable)
}

// Get returns the memo entry associated with the given key.
func (t MapTable) Get(id, pos int) (*Entry, bool) {
	k := key{
		id:  id,
		pos: pos,
	}
	lock.Lock()
	defer lock.Unlock()
	es, ok := t[k]
	if !ok {
		return nil, false
	}

	max := 0
	for i, e := range es[1:] {
		if e.length > es[max].length {
			max = i + 1
		}
	}
	return es[max], ok
}

// Put places a new memoization entry at the given key.
func (t MapTable) Put(id, start, length, examined int, captures []*Capture) {
	k := key{
		id:  id,
		pos: start,
	}
	lock.Lock()
	defer lock.Unlock()
	t[k] = append(t[k], newEntry(id, start, length, examined, captures))
}

// Delete deletes a memoization entry associated with a key.
func (t MapTable) Delete(k key) {
	delete(t, k)
}

// ApplyEdit invalidates memoization entries that overlap with the edit, and
// updates memoization entries that are to the right of the edit.
func (t MapTable) ApplyEdit(e Edit) {
	size := e.Len - (e.End - e.Start)

	for key, ents := range t {
		v := 0 // valid slots
		for _, ent := range ents {
			x1, x2 := key.pos, key.pos+ent.examined
			y1, y2 := e.Start, e.End
			if x1 <= y2 && y1 <= x2 {
				// delete this interval because it overlaps
				continue
			}

			// shift entries that are to the right of the edited interval
			if key.pos >= e.Start && size != 0 {
				// TODO: not sure if this is fully correct
				key.pos += size
				ent.start += size

				// If there is an entry where we want to move to, that is
				// unfortunate we will just invalidate this entry, even though
				// it's not strictly necessary for simplicity (hopefully this
				// does not happen often)
				_, ok := t[key]
				if !ok {
					t[key] = append(t[key], ent)
				}
				continue
			}

			// don't delete this element
			t[key][v] = ent
			v++
		}

		for i := v; i < len(t[key]); i++ {
			t[key][i] = nil
		}
		t[key] = t[key][:v]
		if len(t[key]) == 0 {
			t[key] = nil
			delete(t, key)
		}
	}
}

func (t MapTable) Overlaps(low, high int) []*Entry {
	result := make([]*Entry, 0)
	for key, ents := range t {
		for _, ent := range ents {
			x1, x2 := key.pos, key.pos+ent.examined
			y1, y2 := low, high
			if x1 <= y2 && y1 <= x2 {
				result = append(result, ent)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Start() < result[j].Start()
	})
	return result
}

// Size returns the number of memoization entries in the table.
func (t MapTable) Size() int {
	return len(t)
}
