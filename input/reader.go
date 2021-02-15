package input

import (
	"io"
	"sync"
)

// readerWrapper implements a io.ReaderAt from an io.Reader. The readerWrapper
// works by storing every byte read from the reader, and using that to read
// data that has been read before.
type readerWrapper struct {
	reader io.Reader
	buf    []byte
	lock   sync.Mutex
}

// FromReader converts an io.Reader to an io.ReaderAt.
func FromReader(r io.Reader) io.ReaderAt {
	return &readerWrapper{
		reader: r,
	}
}

// ReadAt implements the io.ReaderAt interface to wrap an io.Reader. Note that
// calls to ReadAt may change the offset within the wrapped io.Reader (since
// Read is called on the wrapped io.Reader to fetch data).
func (r *readerWrapper) ReadAt(b []byte, off int64) (n int, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	blen := int64(len(b))
	// if there is enough space to fill up b already in the buffer, just copy
	// the data and return it.
	if int64(len(r.buf))-off >= blen {
		return copy(b, r.buf[off:]), nil
	}

	// otherwise read data until there is enough or there is an error.
	tmp := make([]byte, bufsz)
	for int64(len(r.buf))-off < blen {
		n, err = r.reader.Read(tmp)
		r.buf = append(r.buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	if off >= int64(len(r.buf)) {
		return 0, err
	}

	return copy(b, r.buf[off:]), err
}

// Slice returns the slice [low:high) in the given ReaderAt.
func Slice(r io.ReaderAt, low, high int) []byte {
	buf := make([]byte, high-low)
	n, _ := r.ReadAt(buf, int64(low))
	return buf[:n]
}
