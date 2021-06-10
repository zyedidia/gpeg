package interval

import "fmt"

type Interval struct {
	Low, High int
}

func (i Interval) String() string {
	return fmt.Sprintf("[%d, %d)", i.Low, i.High)
}

func (i Interval) Shift(amt int) Interval {
	return Interval{
		Low:  i.Low + amt,
		High: i.High + amt,
	}
}

func (i Interval) Len() int {
	return i.High - i.Low
}

func Overlaps(i1, i2 Interval) bool {
	return i1.Low < i2.High && i1.High > i2.Low
}
