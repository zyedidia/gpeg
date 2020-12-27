package input

// Pos represents a position in the input. For now this is implemented as an
// offset, but in the future this may be an interface to allow other sorts of
// types (e.g., line/col).
type Pos struct {
	Off int
}

// PosFromOff returns a new position given a character offset.
func PosFromOff(off int) Pos {
	return Pos{
		Off: off,
	}
}

// Move shifts this position by the desired amount.
func (p Pos) Move(amt int) Pos {
	return Pos{
		Off: p.Off + amt,
	}
}

// Cmp compares this position to another, and returns the number of bytes
// between the two positions.
// For p > other: returns a positive number.
// For p == other: returns 0.
// For p < other: returns a negative number.
func (p Pos) Cmp(other Pos) int {
	return p.Off - other.Off
}
