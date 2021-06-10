package memo

// A Table is an interface for a memoization table data structure. The
// memoization table tracks memoized parse results corresponding to a
// non-terminal parsed at a certain location. The table interface defines the
// ApplyEdit function which is crucial for incremental parsing.
type Table interface {
	// Get returns the entry associated with the given position and ID. If
	// there are multiple entries with the same ID at that position, the
	// largest entry is returned (determined by matched length).
	Get(id, pos int) (*Entry, bool)

	// Put adds a new entry to the table.
	Put(id, start, length, examined, count int, captures []*Capture)

	// ApplyEdit updates the table as necessary when an edit occurs. This
	// operation invalidates all entries within the range of the edit and
	// shifts entries that are to the right of the edit as necessary.
	ApplyEdit(Edit)

	// Overlaps returns all entries that overlap with the given interval
	// [low:high) where the interval of an entry is defined as
	// [start:start+examined).
	// Overlaps(low, high int) []*Entry

	// Size returns the number of entries in the table.
	Size() int
}
