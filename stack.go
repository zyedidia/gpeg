package gpeg

type stack struct {
	entries []stackEntry
}

func NewStack() *stack {
	return &stack{
		entries: []stackEntry{},
	}
}

func (s *stack) Push(entry stackEntry) {
	s.entries = append(s.entries, entry)
}

func (s *stack) Pop() (stackEntry, bool) {
	if len(s.entries) == 0 {
		return stackEntry{}, false
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return ret, true
}

func (s *stack) Peek() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}

func (s *stack) BacktrackEntry(ip, off int, caplist []capture) stackEntry {
	return stackEntry{
		raddress: -1,
		btrack: backtrackEntry{
			ip:      ip,
			off:     off,
			caplist: caplist,
		},
	}
}

func (s *stack) ReturnAddressEntry(addr int) stackEntry {
	return stackEntry{
		raddress: addr,
	}
}

type stackEntry struct {
	// if raddress is -1 use the backtrackEntry
	raddress int
	btrack   backtrackEntry
}

func (se stackEntry) ReturnAddress() bool {
	return se.raddress != -1
}

type backtrackEntry struct {
	ip      int
	off     int
	caplist []capture
}
