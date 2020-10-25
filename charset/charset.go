package charset

import (
	"math/bits"
)

const log2WordSize = 6
const wordSize = 64

type Set struct {
	Bits [2]uint64
}

// New returns a charset which accepts all chars in `chars`. Note
// that all chars must be valid ASCII characters (<128).
func New(chars []byte) Set {
	var set Set
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
func Range(low, high byte) Set {
	var set Set
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
func (c Set) Complement() Set {
	return Set{
		Bits: [2]uint64{^c.Bits[0], ^c.Bits[1]},
	}
}

// Add combines the characters two charsets match together.
func (c Set) Add(c1 Set) Set {
	return Set{
		Bits: [2]uint64{c1.Bits[0] | c.Bits[0], c1.Bits[1] | c.Bits[1]},
	}
}

func (c Set) Sub(c1 Set) Set {
	return Set{
		Bits: [2]uint64{^c1.Bits[0] & c.Bits[0], ^c1.Bits[1] & c.Bits[1]},
	}
}

// Size returns the number of chars matched by this Set.
func (c Set) Size() int {
	return bits.OnesCount64(c.Bits[0]) + bits.OnesCount64(c.Bits[1])
}

// Has checks if a charset accepts a character.
// Pointer receiver is for performance.
func (c *Set) Has(r byte) bool {
	return r < 128 && c.Bits[r>>log2WordSize]&(uint64(1)<<(r&(wordSize-1))) != 0
}

// String returns the string representation of the charset.
func (c Set) String() string {
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
