package memo

type NoneTable struct{}

func (t NoneTable) Get(k Key) (Entry, bool) {
	return Entry{}, false
}

func (t NoneTable) Put(k Key, e Entry) {}
func (t NoneTable) Delete(k Key)       {}
func (t NoneTable) ApplyEdit(e Edit)   {}
func (t NoneTable) Size() int          { return 0 }
