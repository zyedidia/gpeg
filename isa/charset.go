package isa

import "math/bits"

const log2WordSize = 6
const wordSize = 64

type Charset struct {
	Bits [2]uint64
}

// NewCharset returns a charset which accepts all chars in `chars`. Note
// that all chars must be valid ASCII characters (<128).
func NewCharset(chars []byte) Charset {
	var set Charset
	for _, r := range chars {
		switch {
		case r < 64:
			bit := uint64(1) << r
			set.Bits[0] |= bit
		case r < 128:
			bit := uint64(1) << (r - 64)
			set.Bits[1] |= bit
		}
	}

	return set
}

// CharsetRange returns a charset matching all characters between `low` and
// `high` inclusive.
func CharsetRange(low, high byte) Charset {
	var set Charset
	for c := low; c <= high; c++ {
		switch {
		case c < 64:
			bit := uint64(1) << c
			set.Bits[0] |= bit
		case c < 128:
			bit := uint64(1) << (c - 64)
			set.Bits[1] |= bit
		}
	}

	return set
}

// Complement returns a charset that matches all characters except for those
// matched by `c`.
func (c Charset) Complement() Charset {
	return Charset{
		Bits: [2]uint64{^c.Bits[0], ^c.Bits[1]},
	}
}

// Add combines the characters two charsets match together.
func (c Charset) Add(c1 Charset) Charset {
	return Charset{
		Bits: [2]uint64{c1.Bits[0] | c.Bits[0], c1.Bits[1] | c.Bits[1]},
	}
}

// String returns the string representation of the charset.
func (c Charset) String() string {
	s := ""
	inRange := false
	for b := byte(0); b <= 128; b++ {
		if c.Has(b) && !inRange {
			inRange = true
			if c.Has(b + 1) {
				s += "'" + string(b) + "'.."
			}
		} else if !c.Has(b) && inRange {
			inRange = false
			s += "'" + string(b-1) + "',"
		}
	}
	if s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	s = "{" + s + "}"
	return s
}

// Size returns the number of chars matched by this Charset.
func (c Charset) Size() int {
	return bits.OnesCount64(c.Bits[0]) + bits.OnesCount64(c.Bits[1])
}

// Has checks if a charset accepts a character.
// Pointer receiver is for performance.
func (c *Charset) Has(r byte) bool {
	return r < 128 && c.Bits[r>>log2WordSize]&(uint64(1)<<(r&(wordSize-1))) != 0
}
