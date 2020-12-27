package memo

// NoneTable implements a memoization table that does nothing.
type NoneTable struct{}

// Get always returns 'not found'
func (t NoneTable) Get(k Key) (*Entry, bool) {
	return nil, false
}

func (t NoneTable) Put(k Key, e *Entry) {}
func (t NoneTable) Delete(k Key)        {}
func (t NoneTable) ApplyEdit(e Edit)    {}
func (t NoneTable) Size() int           { return 0 }
