package input

// Pos represents a position in the reader in question. This might be
// simply a string offset, or something more complex like a line and
// column number or other representation.
type Pos interface {
	// Less returns true if this position is less than p.
	Less(p Pos) bool
	// GreaterEqual returns true if this position is greater than or equal
	// to p.
	GreaterEqual(p Pos) bool
	// Distance returns the number of bytes between this position and p.
	Distance(p Pos) int
	// Advance moves this position forward byte n bytes.
	Advance(n int) Pos
}

// A Reader is an interface for reading bytes in chunks from a document
// that may have a more complex position representation.
type Reader interface {
	// ReadAtPos reads up to len(b) bytes into b starting at the
	// position p. It returns the number of bytes read, and a possible
	// error. It is perfectly valid for the function to return
	// n < len(b) bytes and a nil error. For efficiency, the reader
	// should try to read as many bytes as possible into b.
	ReadAtPos(b []byte, p Pos) (n int, err error)
}

// A BufferedReader is an efficient wrapper of a Reader which provides
// a nicer interface and avoids repeated interface function calls and
// uses a cache for buffered reading.
type BufferedReader struct {
	r     Reader
	base  Pos
	end   Pos
	off   int
	chunk [4096]byte
	size  int
}

// NewBufferedReader returns a new buffered reader from a general reader
// at the given starting position.
func NewBufferedReader(r Reader, start Pos) *BufferedReader {
	br := BufferedReader{
		r:    r,
		base: start,
		off:  0,
	}
	// TODO: instead of copying just copy the slice/pointer
	br.size, _ = r.ReadAtPos(br.chunk[:], start)
	br.end = br.base.Advance(br.size)
	return &br
}

// Peek returns the next byte but does not consume it.
func (br *BufferedReader) Peek() byte {
	return br.chunk[br.off]
}

// SeekTo moves the reader to a new position.
func (br *BufferedReader) SeekTo(pos Pos) error {
	var err error
	// if the requested position lies inside the current chunk just adjust
	// the offset
	if pos.GreaterEqual(br.base) && pos.Less(br.end) {
		br.off = br.base.Distance(pos)
		return nil
	}

	// otherwise we need to read a new chunk
	br.base = pos
	br.off = 0
	br.size, err = br.r.ReadAtPos(br.chunk[:], br.base)
	br.end = br.base.Advance(br.size)
	return err
}

// Advance moves the reader forward from its current position by n bytes.
func (br *BufferedReader) Advance(n int) error {
	br.off += n

	if br.off >= br.size {
		return br.SeekTo(br.end.Advance(br.off - br.size))
	}
	return nil
}

// Offset returns the current position of the reader.
func (br *BufferedReader) Offset() Pos {
	return br.base.Advance(br.off)
}
