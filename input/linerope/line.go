package linerope

import (
	"bytes"
)

type loc struct {
	line int
	col  int
}

var lzero = loc{0, 0}

func llen(b, sep []byte) loc {
	lines := bytes.Count(b, sep)

	if lines != 0 {
		last := bytes.LastIndex(b, sep) + len(sep)
		return loc{
			line: lines,
			col:  len(b) - last,
		}
	}
	return loc{
		line: 0,
		col:  len(b),
	}
}

func addlocs(a, b loc) loc {
	if a.line != 0 && b.line != 0 {
		return loc{
			line: a.line + b.line,
			col:  b.col,
		}
	} else if a.line != 0 {
		return loc{
			line: a.line,
			col:  b.col + a.col,
		}
	} else if b.line != 0 {
		return loc{
			line: b.line,
			col:  b.col,
		}
	}
	return loc{
		line: 0,
		col:  a.col + b.col,
	}
}

func sublocs(a, b loc) loc {
	if a.line == b.line {
		return loc{
			line: a.line - b.line,
			col:  a.col - b.col,
		}
	}

	return loc{
		line: a.line - b.line,
		col:  a.col,
	}
}

func (l loc) cmp(other loc) int {
	if l.line == other.line {
		if l.col < other.col {
			return -1
		} else if l.col > other.col {
			return 1
		}
		return 0
	} else if l.line < other.line {
		return -1
	}
	return 1
}

func minloc(a, b loc) loc {
	if a.cmp(b) < 0 {
		return a
	}
	return b
}

func maxloc(a, b loc) loc {
	if a.cmp(b) > 0 {
		return a
	}
	return b
}

func sliceloc(b, sep []byte, start, end loc) []byte {
	soff := indexN(b, sep, start.line) + len(sep) + start.col
	eoff := indexN(b, sep, end.line) + len(sep) + end.col
	return b[soff:eoff]
}
