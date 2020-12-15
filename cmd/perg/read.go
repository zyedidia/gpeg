package main

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
	. "github.com/golang-collections/go-datastructures/augmentedtree"
)

type Line struct {
	Start, End int
	Num        int
}

func (l Line) LowAtDimension(uint64) int64 {
	return int64(l.Start)
}

func (l Line) HighAtDimension(uint64) int64 {
	return int64(l.End)
}

func (l Line) OverlapsAtDimension(i Interval, _ uint64) bool {
	x1, x2 := int64(l.Start), int64(l.End)
	y1, y2 := i.LowAtDimension(0), l.HighAtDimension(0)
	return x1 <= y2 && y1 <= x2
}

func (l Line) ID() uint64 {
	return uint64(l.Num)
}

var lineSep = []byte{'\n'}

// TODO: support CRLF and line ending auto-detection
// TODO: optimize by guessing number of lines

// lfs is guaranteed to contain enough space for all the newline indices
func getNewlines(b []byte) []Interval {
	nlines := bytes.Count(b, lineSep)
	lines := make([]Interval, 0, nlines)

	acc := 0
	idx := bytes.Index(b, lineSep)
	last := 0

	n := 0
	for idx != -1 {
		cur := acc + idx
		lines = append(lines, Line{
			Start: last,
			End:   cur,
			Num:   n,
		})
		last = cur + len(lineSep)
		acc += idx + len(lineSep)
		idx = bytes.Index(b[acc:], lineSep)
		n++
	}

	return lines
}

func readMmap(f *os.File) ([]byte, []Interval, error) {
	b, err := mmap.Map(f, mmap.COPY, 0)
	if err != nil {
		return nil, nil, err
	}

	lines := getNewlines(b)
	return b, lines, nil
}

const bufsz = 4096 * 4

func readGeneral(r io.Reader, sz int64) ([]byte, []Interval, error) {
	b := make([]byte, sz)
	lineidx := make([]int, 1)
	buf := make([]byte, bufsz)
	next := 0

	br := bufio.NewReader(r)

	tot := 0
	for {
		n, err := br.Read(buf)

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		acc := 0
		idx := bytes.Index(buf[:n], lineSep)
		for idx != -1 {
			lineidx = append(lineidx, tot+idx+len(lineSep))
			acc += idx + len(lineSep)
			tot += idx + len(lineSep)
			idx = bytes.Index(buf[acc:n], lineSep)
		}
		tot += n - acc

		copy(b[next:], buf[:n])
		next += n
	}

	lines := make([]Interval, 0, len(lineidx)-1)
	for i := 1; i < len(lineidx); i++ {
		end := lineidx[i] - len(lineSep)
		lines = append(lines, Line{
			Start: lineidx[i-1],
			End:   end,
			Num:   i - 1,
		})
	}

	return b, lines, nil
}

func read(r io.Reader, sz int64) ([]byte, []Interval, error) {
	if f, ok := r.(*os.File); ok {
		return readMmap(f)
	} else {
		return readGeneral(r, sz)
	}
}
