package memo

import "github.com/zyedidia/gpeg/memo/shifti"

// ITreeTable implements a memoization table using an interval tree (augmented
// to support efficient shifting).
type ITreeTable struct {
	*shifti.Tree
}

func NewITreeTable() *ITreeTable {
	return &ITreeTable{
		Tree: shifti.NewTree(1),
	}
}

func (t *ITreeTable) Get(k Key) (*Entry, bool) {
	entries := t.Tree.Query(&Entry{
		start:    k.Pos,
		examined: 1,
	})
	for _, e := range entries {
		switch e := e.(type) {
		case *Entry:
			if e.id == k.Id && e.start == k.Pos {
				return e, true
			}
		}
	}
	return nil, false
}

func (t *ITreeTable) Put(k Key, e *Entry) {
	t.Tree.Add(e)
}

func (t *ITreeTable) Delete(k Key) {
}

func (t *ITreeTable) ApplyEdit(e Edit) {
	entries := t.Tree.Query(&Entry{
		start:    e.Start,
		examined: e.End.Off - e.Start.Off,
	})

	for _, e := range entries {
		t.Tree.Delete(e)
	}

	t.Tree.Shift(0, int64(e.Start.Off), int64((e.End.Off-e.Start.Off)+e.Len))
}

func (t *ITreeTable) Size() int {
	return int(t.Tree.Len())
}
