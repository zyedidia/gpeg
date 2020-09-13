package gpeg

type charset struct {
	ascii      [2]uint64
	other      []rune
	complement bool
}

func Charset(chars []rune) charset {
	set := charset{
		other: []rune{},
	}
	for _, r := range chars {
		switch {
		case r < 64:
			bit := uint64(1) << uint32(r)
			set.ascii[0] |= bit
		case r < 128:
			bit := uint64(1) << uint32(r-64)
			set.ascii[1] |= bit
		default:
			set.other = append(set.other, r)
		}
	}

	return set
}

func CharsetRange(low, high rune) charset {
	chars := make([]rune, 0, high-low+1)
	for r := low; r <= high; r++ {
		chars = append(chars, r)
	}
	return Charset(chars)
}

func (c charset) Complement() charset {
	return charset{
		ascii:      [2]uint64{^c.ascii[0], ^c.ascii[1]},
		other:      c.other,
		complement: true,
	}
}

func (c charset) Add(c1 charset) charset {
	return charset{
		ascii: [2]uint64{c1.ascii[0] | c.ascii[0], c1.ascii[1] | c.ascii[1]},
		other: append(c.other, c1.other...),
	}
}

func (c charset) Has(r rune) bool {
	switch {
	case r < 64:
		bit := uint64(1) << uint32(r)
		return c.ascii[0]&bit != 0
	case r < 128:
		bit := uint64(1) << uint32(r-64)
		return c.ascii[1]&bit != 0
	default:
		for _, char := range c.other {
			if r == char {
				return !c.complement
			}
		}
	}
	return c.complement
}
