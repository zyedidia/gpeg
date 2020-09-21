package vm

type stack struct {
	entries []stackEntry
}

func newStack() stack {
	return stack{
		entries: []stackEntry{},
	}
}

func (s stack) push(ent stackEntry) {
	s.entries = append(s.entries, ent)
}

func (s stack) pop() (stackEntry, bool) {
	if len(s.entries) == 0 {
		return stackEntry{}, false
	}

	ret := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return ret, true
}

func (s stack) peek() *stackEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}

type stackEntry struct {
}
