package memo

import "github.com/zyedidia/gpeg/input"

// An Edit represents a modification to the subject string where the interval
// [Start, End) is modified to be Len bytes. If Len = 0, this is equivalent
// to deleting the interval, and if Start = End this is an insertion.
type Edit struct {
	Start, End input.Pos
	Len        int
}

type Table interface {
	// Get returns the entry associated with a given key, and a boolean indicating
	// whether the key exists in the table.
	Get(Key) (*Entry, bool)

	// Put adds a new key-entry pair to the table.
	Put(Key, *Entry)

	// Delete causes the entry associated with the given key to be immediately
	// evicted from the table.
	Delete(Key)

	// ApplyEdit applies the given edit to the memo table by shifting entry
	// locations properly and invaliding any entries in the modified interval.
	ApplyEdit(Edit)

	// Size returns the number of entries in the table.
	Size() int
}
