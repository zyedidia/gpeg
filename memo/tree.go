package memo

import (
	"sync"

	"github.com/zyedidia/gpeg/memo/interval"
)

// TreeTable implements a memoization table using an interval tree (augmented
// to support efficient shifting).
type TreeTable struct {
	*interval.Tree
	threshold int
	lock      sync.Mutex
}

func NewTreeTable(threshold int) *TreeTable {
	return &TreeTable{
		Tree:      &interval.Tree{},
		threshold: threshold,
	}
}

func (t *TreeTable) Get(id, pos int) (*Entry, bool) {
	t.lock.Lock()
	entry := t.Tree.FindLargest(id, pos)
	t.lock.Unlock()
	e, ok := entry.(*Entry)
	return e, ok
}

func (t *TreeTable) Put(id, start, length, examined, count int, captures []*Capture) {
	if examined < t.threshold || length == 0 {
		return
	}

	e := &Entry{
		length:   length,
		count:    count,
		captures: captures,
	}
	t.lock.Lock()
	e.setPos(t.Tree.Add(id, start, start+examined, e))
	t.lock.Unlock()
}

func (t *TreeTable) ApplyEdit(e Edit) {
	low, high := e.Start, e.End
	if low == high {
		high = low + 1
	}
	amt := e.Len - (e.End - e.Start)
	t.Tree.RemoveAndShift(low, high, amt)
}
