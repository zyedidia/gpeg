package input

type Pos int

// A Reader is an interface for reading bytes in chunks from a document
// that may have a more complex position representation.
type Reader interface {
	// ReadAtPos reads as many bytes as possible from p and returns the
	// result as a slice. It is expected that the data being read is already
	// in memory, and as a result the slice that is returned does not cause
	// any allocation apart from the fat pointer for the slice itself.
	ReadAtPos(p Pos) (b []byte, err error)
}

// A BufferedReader is an efficient wrapper of a Reader which provides
// a nicer interface and avoids repeated interface function calls and
// uses a cache for buffered reading.
type BufferedReader struct {
	r     Reader
	base  Pos
	off   int
	chunk []byte
	end   bool
}

var dummy = []byte{0}

// NewBufferedReader returns a new buffered reader from a general reader
// at the given starting position.
func NewBufferedReader(r Reader, start Pos) *BufferedReader {
	br := BufferedReader{
		r:    r,
		base: start,
		off:  0,
	}
	br.readAtBase()
	return &br
}

func (br *BufferedReader) readAtBase() {
	var err error
	br.chunk, err = br.r.ReadAtPos(br.base)
	br.end = len(br.chunk) == 0 || err != nil
	if br.end {
		br.off = 0
		br.chunk = dummy
	}
}

// Peek returns the next byte but does not consume it.
func (br *BufferedReader) Peek() (byte, bool) {
	return br.chunk[br.off], !br.end
}

// SeekTo moves the reader to a new position.
func (br *BufferedReader) SeekTo(pos Pos) bool {
	if pos < br.base+Pos(len(br.chunk)) && pos >= br.base {
		br.off = int(pos - br.base)
		return true
	}

	br.base = pos
	br.off = 0
	br.readAtBase()
	return !br.end
}

// Advance moves the reader forward from its current position by n bytes.
func (br *BufferedReader) Advance(n int) bool {
	br.off += n

	if br.off >= len(br.chunk) {
		return br.SeekTo(br.base + Pos(br.off))
	}
	return true
}

// Offset returns the current position of the reader.
func (br *BufferedReader) Offset() Pos {
	return br.base + Pos(br.off)
}
