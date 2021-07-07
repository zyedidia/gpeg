package lazylog

import "fmt"

type interval struct {
	low, high int
	value     interface{}
}

func (i *interval) Low() int {
	return i.low
}

func (i *interval) High() int {
	return i.high
}

func (i *interval) length() int {
	return i.High() - i.Low()
}

func (i *interval) Shift(amt int) {
	i.low += amt
	i.high += amt
}

func (i *interval) String() string {
	return fmt.Sprintf("[%d, %d)", i.low, i.high)
}

// returns true if i1 overlaps with the interval [low:high)
func overlaps(i1 interval, low, high int) bool {
	return i1.Low() <= high && i1.High() >= low
}
