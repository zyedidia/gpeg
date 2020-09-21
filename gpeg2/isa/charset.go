package isa

type Charset struct {
	ascii [2]uint64
}

// NewCharset returns a charset which accepts all chars in `chars`. Note
// that all chars must be valid ASCII characters (<128).
func NewCharset(chars []byte) Charset {
	var set Charset
	for _, r := range chars {
		switch {
		case r < 64:
			bit := uint64(1) << uint32(r)
			set.ascii[0] |= bit
		case r < 128:
			bit := uint64(1) << uint32(r-64)
			set.ascii[1] |= bit
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
			bit := uint64(1) << uint32(c)
			set.ascii[0] |= bit
		case c < 128:
			bit := uint64(c) << uint32(c-64)
			set.ascii[1] |= bit
		}
	}

	return set
}

// Complement returns a charset that matches all characters except for those
// matched by `c`.
func (c Charset) Complement() Charset {
	return Charset{
		ascii: [2]uint64{^c.ascii[0], ^c.ascii[1]},
	}
}

// Add combines the characters two charsets match together.
func (c Charset) Add(c1 Charset) Charset {
	return Charset{
		ascii: [2]uint64{c1.ascii[0] | c.ascii[0], c1.ascii[1] | c.ascii[1]},
	}
}

// Has checks if a charset accepts a character.
func (c Charset) Has(r byte) bool {
	switch {
	case r < 64:
		bit := uint64(1) << uint32(r)
		return c.ascii[0]&bit != 0
	case r < 128:
		bit := uint64(1) << uint32(r-64)
		return c.ascii[1]&bit != 0
	}
	return false
}
