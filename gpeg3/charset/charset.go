package charset

import (
	"math/bits"
	"strconv"
)

const log2WordSize = 6
const wordSize = 64

// A Set represents a set of chars.
type Set struct {
	// Bits is the bit array for indicating which chars are in the set.
	// We have 256 bits because a char can have 256 different values.
	Bits [4]uint64
}

// A SmallSet is the same as a Set but can only represent 128 possible chars.
// This is an optimization, since in the common case, only ASCII bytes are
// used which are <128. The full Set is only necessary when unicode control
// characters must be matched.
type SmallSet struct {
	Bits [2]uint64
}

// Size returns the number of chars matched by this Set.
func (c SmallSet) Size() int {
	return bits.OnesCount64(c.Bits[0]) + bits.OnesCount64(c.Bits[1])
}

// Has checks if a charset accepts a character.
// Pointer receiver is for performance.
func (c *SmallSet) Has(r byte) bool {
	return c.Bits[r>>log2WordSize]&(uint64(1)<<(r&(wordSize-1))) != 0
}

// IsSmall returns true if this set can be converted to a small set. In other
// words, if this set only matches bytes <128.
func (c Set) IsSmall() bool {
	return c.Bits[2] == 0 && c.Bits[3] == 0
}

// SmallSet converts this Set to a SmallSet.
func (c Set) SmallSet() SmallSet {
	return SmallSet{
		Bits: [2]uint64{c.Bits[0], c.Bits[1]},
	}
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
		case r < 192:
			bit := uint64(1) << (r - 128)
			set.Bits[2] |= bit
		default:
			bit := uint64(1) << (r - 192)
			set.Bits[3] |= bit
		}
	}

	return set
}

// CharsetRange returns a charset matching all characters between `low` and
// `high` inclusive.
func Range(low, high byte) Set {
	var set Set
	for c := int(low); c <= int(high); c++ {
		switch {
		case c < 64:
			bit := uint64(1) << c
			set.Bits[0] |= bit
		case c < 128:
			bit := uint64(1) << (c - 64)
			set.Bits[1] |= bit
		case c < 192:
			bit := uint64(1) << (c - 128)
			set.Bits[2] |= bit
		default:
			bit := uint64(1) << (c - 192)
			set.Bits[3] |= bit
		}
	}

	return set
}

// Complement returns a charset that matches all characters except for those
// matched by `c`.
func (c Set) Complement() Set {
	return Set{
		Bits: [4]uint64{^c.Bits[0], ^c.Bits[1], ^c.Bits[2], ^c.Bits[3]},
	}
}

// Add combines the characters two charsets match together.
func (c Set) Add(c1 Set) Set {
	return Set{
		Bits: [4]uint64{c1.Bits[0] | c.Bits[0], c1.Bits[1] | c.Bits[1], c1.Bits[2] | c.Bits[2], c1.Bits[3] | c.Bits[3]},
	}
}

func (c Set) Sub(c1 Set) Set {
	return Set{
		Bits: [4]uint64{^c1.Bits[0] & c.Bits[0], ^c1.Bits[1] & c.Bits[1], ^c1.Bits[2] & c.Bits[2], ^c1.Bits[3] & c.Bits[3]},
	}
}

// Size returns the number of chars matched by this Set.
func (c Set) Size() int {
	return bits.OnesCount64(c.Bits[0]) + bits.OnesCount64(c.Bits[1]) + bits.OnesCount64(c.Bits[2]) + bits.OnesCount64(c.Bits[3])
}

// Has checks if a charset accepts a character.
// Pointer receiver is for performance.
func (c *Set) Has(r byte) bool {
	return c.Bits[r>>log2WordSize]&(uint64(1)<<(r&(wordSize-1))) != 0
}

// String returns the string representation of the charset.
func (c Set) String() string {
	s := ""
	inRange := false
	for b := int(0); b <= 255; b++ {
		if c.Has(byte(b)) && b == 255 {
			s += strconv.QuoteRune(rune(b))
		} else if c.Has(byte(b)) && !inRange {
			inRange = true
			if c.Has(byte(b + 1)) {
				s += strconv.QuoteRune(rune(b)) + ".."
			}
		} else if !c.Has(byte(b)) && inRange {
			inRange = false
			s += strconv.QuoteRune(rune(b-1)) + ","
		}
	}
	if s != "" && s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	s = "{" + s + "}"
	return s
}
