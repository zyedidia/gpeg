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
