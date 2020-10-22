package memo

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
