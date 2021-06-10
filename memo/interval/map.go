package interval

type Value interface{}

type Pos interface {
	Pos() int
}

// An interval map is a key-value data structure that maps intervals to
// values.  Every value is associated with an interval [low, high) and an id.
// Values may be looked up, added, removed, and queried for overlapping
// intervals. The tree also supports efficient shifting of intervals via
// a lazy shift propagation mechanism.
type Map interface {
	// Returns the value associated with the largest interval at (id, pos).
	FindLargest(id, pos int) Value
	// Adds a new value with 'id' and interval [low, high). Returns a value
	// that can be used to locate the inserted value even after shifts have
	// occurred (you may want to associate the Pos with your value).
	Add(id, low, high int, val Value) Pos
	// Removes all values with intervals that overlap [low, high) and then
	// performs a shift of size amt at idx.
	RemoveAndShift(low, high, amt int)
	// Returns the number of values in the tree.
	Size() int
}
