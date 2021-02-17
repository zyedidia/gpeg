package isa

// A Checker is used so the user can perform additional custom validation of
// parse results. For example, you might want to parse only 8-bit integers by
// matching [0-9]+ and then using a checker to ensure the matched integer is in
// the range 0-256.
type Checker interface {
	Check(b []byte) bool
}

type MapChecker map[string]struct{}

func NewMapChecker(strs []string) MapChecker {
	m := make(map[string]struct{})
	for _, s := range strs {
		m[s] = struct{}{}
	}
	return m
}

func (m MapChecker) Check(b []byte) bool {
	_, ok := m[string(b)]
	return ok
}
