package input

import "io"

const bufsz = 4096

// readerWrapper implements a ReaderAtPos from an io.Reader. The readerWrapper
// works by storing every byte read from the reader, and using that to read
// data that has been read before.
type readerWrapper struct {
	reader io.Reader
	buf    []byte
}

// FromReader converts an io.Reader to a ReaderAtPos.
func FromReader(r io.Reader) ReaderAtPos {
	return &readerWrapper{
		reader: r,
	}
}

// ReadAtPos reads from the given position (Pos must represent an offset).
func (r *readerWrapper) ReadAtPos(p Pos) ([]byte, error) {
	if p.Off < len(r.buf) {
		return r.buf[p.Off:], nil
	}

	b := make([]byte, bufsz)
	for p.Off >= len(r.buf) {
		n, err := r.reader.Read(b)
		if err != nil {
			return nil, err
		}
		r.buf = append(r.buf, b[:n]...)
	}
	return r.buf[p.Off:], nil
}

// A readerAtWrapper wraps an io.ReaderAt to create a ReaderAtPos.
type readerAtWrapper struct {
	readerAt io.ReaderAt
	buf      [bufsz]byte
}

// FromReaderAt converts an io.ReaderAt to a ReaderAtPos.
func FromReaderAt(r io.ReaderAt) ReaderAtPos {
	return &readerAtWrapper{
		readerAt: r,
	}
}

// ReadAtPos reads from the given position (Pos must represent an offset).
func (r *readerAtWrapper) ReadAtPos(p Pos) ([]byte, error) {
	_, err := r.readerAt.ReadAt(r.buf[:], int64(p.Off))
	return r.buf[:], err
}
