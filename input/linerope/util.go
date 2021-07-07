package linerope

import (
	"bytes"
)

// indexN finds the index of n-th sep in b.
func indexN(b, sep []byte, n int) (index int) {
	index, idx, sepLen := 0, -1, len(sep)
	for i := 0; i < n; i++ {
		if idx = bytes.Index(b, sep); idx == -1 {
			break
		}
		b = b[idx+sepLen:]
		index += idx
	}

	if idx == -1 {
		index = -1
	} else {
		index += (n - 1) * sepLen
	}

	return
}

// lineCol converts an absolute position to a line/col pair by scanning b.
func lineCol(b, sep []byte, pos int) loc {
	var line, last int
	for {
		idx := bytes.Index(b[last:], sep)
		if idx < 0 {
			break
		} else if last+idx >= pos {
			return loc{line, pos - last}
		}
		last += idx + len(sep)
		line++
	}
	return loc{line, pos - last}
}
