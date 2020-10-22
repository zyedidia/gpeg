package memo

import "github.com/zyedidia/gpeg/input"

// A Table is an LRU cache for memo.Key and memo.Entry values. Each entry is
// put into the table with an associated key used for looking up the entry.
// The Table has a maximum size, and uses a least-recently-used eviction
// policy when there is not space for a new entry.
type Table struct {
	size     int
	capacity int
	lru      list
	table    map[Key]*node
}

// NewTable returns a new memo.Table with the given capacity.
func NewTable(capacity int) *Table {
	return &Table{
		size:     0,
		capacity: capacity,
		lru:      list{},
		table:    make(map[Key]*node),
	}
}

// Get returns the entry associated with a given key, and a boolean indicating
// whether the key exists in the table.
func (t *Table) Get(k Key) (Entry, bool) {
	if n, ok := t.table[k]; ok {
		t.lru.moveHead(n)
		return n.ent, true
	}
	return Entry{}, false
}

// Put adds a new key-entry pair to the table.
func (t *Table) Put(k Key, e Entry) {
	if n, ok := t.table[k]; ok {
		n.ent = e
		t.lru.moveHead(n)
		return
	}

	if t.size == t.capacity {
		key := t.lru.tail.key
		t.lru.remove(t.lru.tail)
		t.size--
		delete(t.table, key)
	}
	n := &node{
		key: k,
		ent: e,
	}
	t.lru.insertHead(n)
	t.size++
	t.table[k] = n
}

// Delete causes the entry associated with the given key to be immediately
// evicted from the table.
func (t *Table) Delete(k Key) {
	if n, ok := t.table[k]; ok {
		t.lru.remove(n)
		t.size--
		delete(t.table, k)
	}
}

// An Edit represents a modification to the subject string where the interval
// [Start, End) is modified to be Len bytes. If Len = 0, this is equivalent
// to deleting the interval, and if Start = End this is an insertion.
type Edit struct {
	Start, End input.Pos
	Len        int
}

// ApplyEdit applies the given edit to the memo table by shifting entry
// locations properly and invaliding any entries in the modified interval.
func (t *Table) ApplyEdit(e Edit) {
	n := t.lru.head
	size := e.Len - int(e.End-e.Start)
	for n != nil {
		// invalidate entries that overlap with the edited interval
		x1, x2 := int(n.key.Pos), int(n.key.Pos)+n.ent.examined
		y1, y2 := int(e.Start), int(e.End)
		if x1 <= y2 && y1 <= x2 {
			torm := n
			n = n.next
			t.lru.remove(torm)
			t.size--
			delete(t.table, torm.key)
			continue
		}

		// shift entries that are to the right of the edited interval
		if n.key.Pos >= e.Start {
			// Since the key is changed, we need to rehash the value
			delete(t.table, n.key)
			n.key.Pos += input.Pos(size)
			t.table[n.key] = n
		}

		n = n.next
	}
}
