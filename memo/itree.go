package memo

import "github.com/zyedidia/gpeg/memo/avlint"

// ITreeTable implements a memoization table using an interval tree (augmented
// to support efficient shifting).
type ITreeTable struct {
	*avlint.Tree
}

func NewITreeTable() *ITreeTable {
	return &ITreeTable{
		Tree: avlint.NewTree(),
	}
}

func (t *ITreeTable) Get(k Key) (*Entry, bool) {
	entry := t.Tree.Search(avlint.Key{
		Val: k.Pos.Off,
		Id:  int(k.Id),
	})
	e, ok := entry.(*Entry)
	return e, ok
}

func (t *ITreeTable) Put(k Key, e *Entry) {
	t.Tree.Add(avlint.Key{
		Val: k.Pos.Off,
		Id:  int(k.Id),
	}, e)
}

func (t *ITreeTable) Delete(k Key) {
	t.Tree.Remove(avlint.Key{
		Val: k.Pos.Off,
		Id:  int(k.Id),
	})
}

func (t *ITreeTable) ApplyEdit(e Edit) {
	entries := t.Tree.Overlap(&Entry{
		start:    e.Start,
		examined: e.End.Off - e.Start.Off,
	})

	for _, e := range entries {
		switch e := e.(type) {
		case *Entry:
			t.Tree.Remove(avlint.Key{
				Val: e.start.Off,
				Id:  int(e.id),
			})
		}
	}

	t.Tree.Shift(e.Start.Off, (e.End.Off-e.Start.Off)+e.Len)
}

func (t *ITreeTable) Size() int {
	return t.Tree.Height()
}
