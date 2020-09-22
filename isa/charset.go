package isa

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

// Has checks if a charset accepts a character.
func (c *Charset) Has(r byte) bool {
	switch {
	case r < 64:
		bit := uint64(1) << r
		return c.Bits[0]&bit != 0
	case r < 128:
		bit := uint64(1) << (r - 64)
		return c.Bits[1]&bit != 0
	}
	return false
}
