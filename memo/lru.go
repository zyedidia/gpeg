package memo

import "github.com/zyedidia/gpeg/input"

// TODO: use a memory pool for the objects

// Doubly linked list for keeping track of the least recently used entry.
type list struct {
	head, tail *node
}

type node struct {
	key        Key
	ent        Entry
	prev, next *node
}

func (l *list) insertHead(n *node) {
	n.next = l.head
	n.prev = nil
	if l.head != nil {
		l.head.prev = n
	} else {
		l.tail = n
	}
	l.head = n
}

func (l *list) moveHead(n *node) {
	l.remove(n)
	l.insertHead(n)
}

func (l *list) remove(n *node) {
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}
}

// A LRUTable is an LRU cache for memo.Key and memo.Entry values. Each entry is
// put into the table with an associated key used for looking up the entry.
// The LRUTable has a maximum size, and uses a least-recently-used eviction
// policy when there is not space for a new entry.
type LRUTable struct {
	size     int
	capacity int
	lru      list
	table    map[Key]*node
}

// NewTable returns a new memo.LRUTable with the given capacity.
func NewLRUTable(capacity int) *LRUTable {
	return &LRUTable{
		size:     0,
		capacity: capacity,
		lru:      list{},
		table:    make(map[Key]*node),
	}
}

// Get returns the entry associated with a given key, and a boolean indicating
// whether the key exists in the table.
func (t *LRUTable) Get(k Key) (Entry, bool) {
	if n, ok := t.table[k]; ok {
		t.lru.moveHead(n)
		return n.ent, true
	}
	return Entry{}, false
}

// Put adds a new key-entry pair to the table.
func (t *LRUTable) Put(k Key, e Entry) {
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
func (t *LRUTable) Delete(k Key) {
	if n, ok := t.table[k]; ok {
		t.lru.remove(n)
		t.size--
		delete(t.table, k)
	}
}

// ApplyEdit applies the given edit to the memo table by shifting entry
// locations properly and invaliding any entries in the modified interval.
func (t *LRUTable) ApplyEdit(e Edit) {
	n := t.lru.head
	size := e.Len - int(e.End-e.Start)
	for n != nil {
		// fmt.Println(n.key.Id, n.key.Pos, n.ent.MatchLength(), n.ent.Examined())
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

// Size returns the number of entries in the table.
func (t *LRUTable) Size() int {
	return t.size
}
