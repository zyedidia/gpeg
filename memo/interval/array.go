package interval

// An Array is another implementation of the interval.Set backed by an array
// rather than an AVL tree. This implementation is naive and ineffecient, but
// provides a good point of comparison for benchmarking and testing.
type Array struct {
	slots []slot
}

type slot struct {
	*ivalue
	id int
}

func (iv *ivalue) Pos() int {
	return iv.interval.Low
}

func (a *Array) FindLargest(id, pos int) Value {
	var max int
	maxi := -1
	for i, in := range a.slots {
		if in.interval.Low == pos && in.id == id && in.interval.High > max {
			maxi = i
			max = in.interval.High
		}
	}
	if maxi == -1 || maxi >= len(a.slots) {
		return nil
	}

	return a.slots[maxi].value
}

func (a *Array) Add(id, low, high int, val Value) Pos {
	iv := &ivalue{
		interval: Interval{low, high},
		value:    val,
	}
	a.slots = append(a.slots, slot{
		id:     id,
		ivalue: iv,
	})
	return iv
}

func (a *Array) RemoveAndShift(low, high, amt int) {
	for i := 0; i < len(a.slots); {
		if Overlaps(a.slots[i].interval, Interval{low, high}) {
			a.slots[i] = a.slots[len(a.slots)-1]
			a.slots = a.slots[:len(a.slots)-1]
		} else {
			i++
		}
	}

	if amt == 0 {
		return
	}

	for i := range a.slots {
		if a.slots[i].interval.Low >= low {
			a.slots[i].interval = a.slots[i].interval.Shift(amt)
		}
	}
}

func (a *Array) Size() int {
	return len(a.slots)
}
