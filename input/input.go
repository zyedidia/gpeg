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

	// Slice returns the bytes in the data [low:high). This is used for
	// resolving capture data.
	Slice(low, high Pos) (b []byte)
}

// A BufferedReader is an efficient wrapper of a Reader which provides
// a nicer interface and avoids repeated interface function calls and
// uses a cache for buffered reading.
type BufferedReader struct {
	r     Reader
	base  Pos
	off   int
	max   Pos
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

func (br *BufferedReader) Reset(r Reader, start Pos) {
	br.max = 0
	br.r = r
	br.base = start
	br.off = 0
	br.end = false
	br.readAtBase()
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

func (br *BufferedReader) MaxExaminedPos() Pos {
	return br.max
}

func (br *BufferedReader) ResetMaxExamined() {
	br.max = 0
}

// Peek returns the next byte but does not consume it.
func (br *BufferedReader) Peek() (byte, bool) {
	if br.base+Pos(br.off) > br.max {
		br.max = br.base + Pos(br.off)
	}
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

// Slice returns the bytes in the data [low:high). This is used for
// resolving capture data.
func (br *BufferedReader) Slice(low, high Pos) []byte {
	return br.r.Slice(low, high)
}

// Offset returns the current position of the reader.
func (br *BufferedReader) Offset() Pos {
	return br.base + Pos(br.off)
}
