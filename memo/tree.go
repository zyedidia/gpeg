package memo

import (
	"sort"
	"sync"

	"github.com/zyedidia/gpeg/memo/avlint"
)

// TreeTable implements a memoization table using an interval tree (augmented
// to support efficient shifting).
type TreeTable struct {
	*avlint.Tree
	threshold int
	lock      sync.Mutex
}

func NewTreeTable(threshold int) *TreeTable {
	return &TreeTable{
		Tree:      &avlint.Tree{},
		threshold: threshold,
	}
}

func (t *TreeTable) Get(id, pos int) (*Entry, bool) {
	t.lock.Lock()
	entry := t.Tree.Search(pos, id)
	t.lock.Unlock()
	e, ok := entry.(*Entry)
	return e, ok
}

func (t *TreeTable) Put(id, start, length, examined, count int, captures []*Capture) {
	if examined < t.threshold || length == 0 {
		return
	}

	e := newEntry(id, start, length, examined, count, captures)
	t.lock.Lock()
	loc := t.Tree.Add(start, start+examined, e, id)
	e.loc = loc
	t.lock.Unlock()
}

func (t *TreeTable) ApplyEdit(e Edit) {
	entries := t.Tree.Overlap(e.Start, e.End)

	for _, e := range entries {
		switch e := e.(type) {
		case *Entry:
			t.Tree.Remove(e.start, int(e.id))
		}
	}

	t.Tree.Shift(e.Start, (e.Start-e.End)+e.Len)
}

func (t *TreeTable) Overlaps(low, high int) []*Entry {
	result := make([]*Entry, 0)
	entries := t.Tree.Overlap(low, high)
	for _, e := range entries {
		result = append(result, e.(*Entry))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Start() == result[j].Start() {
			return result[i].Examined() > result[j].Examined()
		}
		return result[i].Start() < result[j].Start()
	})
	return result
}
