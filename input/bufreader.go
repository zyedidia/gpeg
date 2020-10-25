package input

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

// NewBufferedReader returns a new buffered reader from a general reader.
func NewBufferedReader(r Reader) *BufferedReader {
	br := BufferedReader{
		r:   r,
		off: 0,
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
