package avlint

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"
)

func TestOverlap(t *testing.T) {
	tests := []struct {
		i1, i2  interval
		overlap bool
	}{
		{
			interval{low: 0, high: 10},
			interval{low: 1, high: 2},
			true,
		},
		{
			interval{low: 0, high: 10},
			interval{low: 10, high: 15},
			false,
		},
		{
			interval{low: 0, high: 10},
			interval{low: 9, high: 15},
			true,
		},
		{
			interval{low: 0, high: 10},
			interval{low: -10, high: 15},
			true,
		},
		{
			interval{low: 0, high: 10},
			interval{low: -10, high: 0},
			false,
		},
		{
			interval{low: 0, high: 10},
			interval{low: 10, high: 11},
			false,
		},
		{
			interval{low: 0, high: 10},
			interval{low: 0, high: 1},
			true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("overlap%d", i), func(t *testing.T) {
			got := overlaps(&tt.i1, tt.i2.low, tt.i2.high)
			if got != tt.overlap {
				t.Errorf("incorrect overlap between %v, %v: want %v, got %v", &tt.i1, &tt.i2, tt.overlap, got)
			}
		})
	}
}

// var seed = time.Now().UTC().UnixNano()

func TestMain(m *testing.M) {
	// rand.Seed(seed)
	os.Exit(m.Run())
}

// TestTreeMap tests the interval tree as a set of intervals. It uses a map as
// the expected behavior for the tree. Inserts, Deletes, and Searches are
// performed at random and each search is checked for the correct result.
func TestTreeMap(t *testing.T) {
	tree := &Tree{}
	m := make(map[key]Interval)

	const (
		opAdd = iota
		opRemove
		opSearch

		maxPos = 10
		nops   = 10000
	)

	ids := 1
	for i := 0; i < nops; i++ {
		op := rand.Intn(3)
		pos := rand.Intn(maxPos)
		id := rand.Intn(ids)
		switch op {
		case opAdd:
			in := &interval{
				low:  pos,
				high: pos + rand.Intn(20),
			}
			tree.Add(in, ids)
			m[key{
				pos: pos,
				id:  ids,
			}] = in
			ids++
		case opRemove:
			tree.Remove(pos, id)
			delete(m, key{
				pos: pos,
				id:  id,
			})
		case opSearch:
			tin := tree.Search(pos, id)
			min := m[key{
				pos: pos,
				id:  id,
			}]
			if min == nil {
				if tin != nil {
					t.Errorf("incorrect search: want %v, got %v", min, tin)
				}
				continue
			}
			if tin.Low() != min.Low() || tin.High() != min.High() {
				t.Errorf("incorrect search: want %v, got %v", min, tin)
			}
		}

		if len(m) != tree.Size() {
			t.Errorf("incorrect size: want %d, got %d", len(m), tree.Size())
		}
	}
}

type intervalArray []interval

func (a intervalArray) Overlap(low, high int) []Interval {
	result := make([]Interval, 0)
	for i := range a {
		if overlaps(&a[i], low, high) {
			result = append(result, &a[i])
		}
	}
	return result
}

func (a intervalArray) Shift(idx, amt int) {
	for i := range a {
		if a[i].low >= idx {
			a[i].low += amt
			a[i].high += amt
		}
	}
}

func (a intervalArray) Find(pos, id int) int {
	for i, in := range a {
		if in.low == pos && in.id == id {
			return i
		}
	}
	return -1
}

// TestOverlapShift tests the Overlap and Shift operations.
func TestTreeOverlapShift(t *testing.T) {
	tree := &Tree{}
	array := make(intervalArray, 0)

	const (
		opAdd = iota
		opShift
		opOverlap

		nops   = 100000
		maxPos = 1000
	)

	for i := 0; i < nops; i++ {
		op := rand.Intn(3)
		switch op {
		case opAdd:
			low := rand.Intn(maxPos)
			high := low + rand.Intn(100) + 1
			array = append(array, interval{
				low:  low,
				high: high,
				id:   i,
			})
			tree.Add(&interval{
				low:  low,
				high: high,
				id:   i,
			}, i)
		case opShift:
			idx := rand.Intn(maxPos)
			amt := randrange(max(-idx, -20), 20)
			var tins []Interval
			if amt < 0 {
				tins = tree.Overlap(idx, idx-amt)
			} else {
				tins = tree.Overlap(idx, idx+1)
			}

			for _, in := range tins {
				intrvl := in.(*interval)
				tree.Remove(intrvl.Low(), intrvl.id)

				i := array.Find(intrvl.Low(), intrvl.id)
				if i != -1 {
					array = append(array[:i], array[i+1:]...)
				}
			}

			tree.Shift(idx, amt)
			array.Shift(idx, amt)
		case opOverlap:
			low := rand.Intn(maxPos)
			high := low + rand.Intn(100)

			tins := tree.Overlap(low, high)
			ains := array.Overlap(low, high)

			sortIntervals(tins)
			sortIntervals(ains)

			if len(tins) != len(ains) {
				t.Errorf("incorrect length: want %d, got %d", len(ains), len(tins))
				os.Exit(1)
				continue
			}

			for i := range tins {
				if tins[i].Low() != ains[i].Low() || tins[i].High() != ains[i].High() {
					t.Errorf("%d: incorrect interval: want %v, got %v", i, ains[i], tins[i])
				}
			}
		}
	}
}

func randrange(min, max int) int {
	return rand.Intn(max-min) + min
}

func sortIntervals(r []Interval) {
	sort.Slice(r, func(i, j int) bool {
		if r[i].Low() < r[j].Low() {
			return true
		} else if r[i].Low() > r[j].Low() {
			return false
		}
		return r[i].High() < r[j].High()
	})
}

func TestLocator(t *testing.T) {
	tree := &Tree{}
	i := &interval{
		low:  0,
		high: 10,
	}
	tree.Add(i, 0)
	tree.Shift(0, 5)
	start := i.loc.Start()
	if start != 5 {
		t.Errorf("incorrect locator: want %d, got %d", 5, start)
	}
}
