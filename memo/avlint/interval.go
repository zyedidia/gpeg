package avlint

type Interval interface {
	Low() int
	High() int
	ShiftLow(amt int)
	ShiftHigh(amt int)
	Overlaps(i Interval) bool
}
