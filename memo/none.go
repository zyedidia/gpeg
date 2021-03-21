package memo

// NoneTable implements a memoization table that does nothing.
type NoneTable struct{}

// Get always returns 'not found'
func (t NoneTable) Get(id, pos int) (*Entry, bool) {
	return nil, false
}

func (t NoneTable) Put(id, start, length, examined, count int, captures []*Capture) {}
func (t NoneTable) ApplyEdit(e Edit)                                                {}
func (t NoneTable) Overlaps(low, high int) []*Entry                                 { return nil }
func (t NoneTable) Size() int                                                       { return 0 }
