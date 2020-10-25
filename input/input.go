package input

// A Reader is an interface for reading bytes in chunks from a document
// that may have a more complex position representation.
type Reader interface {
	// ReadAtPos reads as many bytes as possible from p and returns the
	// result as a slice. It is expected that the data being read is already
	// in memory, and as a result the slice that is returned does not cause
	// any allocation apart from the fat pointer for the slice itself.
	ReadAtPos(p Pos) (b []byte, err error)

	// Slice returns the bytes in the data [low:high). This is used for
	// resolving capture data.
	Slice(low, high Pos) (b []byte)
}
