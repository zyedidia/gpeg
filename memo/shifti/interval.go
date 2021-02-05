package shifti

import "sync"

type Interval interface {
	Low(dim uint64) int64
	ShiftLow(dim uint64, count int64)

	High(dim uint64) int64
	ShiftHigh(dim uint64, count int64)

	Overlaps(i Interval, dim uint64) bool

	Id() uint64
}

type Intervals []Interval

var iPool = sync.Pool{
	New: func() interface{} {
		return make(Intervals, 0, 10)
	},
}

func (ivs *Intervals) Dispose() {
	for i := 0; i < len(*ivs); i++ {
		(*ivs)[i] = nil
	}
	*ivs = (*ivs)[:0]
	iPool.Put(*ivs)
}
