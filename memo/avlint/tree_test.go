package avlint_test

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/zyedidia/avlint"
)

var seed = time.Now().UTC().UnixNano()

func TestMain(m *testing.M) {
	rand.Seed(seed)
	os.Exit(m.Run())
}

type basicRange struct {
	low, high int
}

func (r *basicRange) Low() int {
	return r.low
}

func (r *basicRange) High() int {
	return r.high
}

func (r *basicRange) ShiftLow(amt int) {
	r.low += amt
}

func (r *basicRange) ShiftHigh(amt int) {
	r.high += amt
}

func (r *basicRange) Overlaps(i avlint.Interval) bool {
	return r.Low() < i.High() && r.High() > i.Low()
}

func (r *basicRange) String() string {
	return fmt.Sprintf("[%d, %d)", r.low, r.high)
}

func TestAdd(t *testing.T) {
	tree := &avlint.Tree{}
	m := make(map[avlint.Key]avlint.Interval)

	const nadd = 100000
	for i := 0; i < nadd; i++ {
		key := avlint.Key{
			Val: rand.Int(),
			Id:  i,
		}
		low := rand.Int()
		value := &basicRange{
			low:  low,
			high: low + 10,
		}
		tree.Add(key, value)
		m[key] = value
	}

	for k := range m {
		tv := tree.Search(k)
		mv := m[k]
		if mv != tv {
			t.Errorf("map and tree disagree: [%d], %d, %d", k, mv, tv)
		}
	}
}

type intervalArray []avlint.Interval

func (a intervalArray) Overlap(i avlint.Interval) []avlint.Interval {
	result := make([]avlint.Interval, 0)
	for _, in := range a {
		if in.Overlaps(i) {
			result = append(result, in)
		}
	}
	return result
}

func sortIntervals(r []avlint.Interval) {
	sort.Slice(r, func(i, j int) bool {
		if r[i].Low() < r[j].Low() {
			return true
		} else if r[i].Low() > r[j].Low() {
			return false
		}
		return r[i].High() < r[j].High()
	})
}

func TestOverlapShift(t *testing.T) {
	tree := &avlint.Tree{}
	array := make(intervalArray, 0)

	const ninterval = 100000
	const max = 100000
	for i := 0; i < ninterval; i++ {
		start := rand.Intn(max)
		end := start + rand.Intn(max)
		tree.Add(avlint.Key{
			Val: start,
			Id:  i,
		}, &basicRange{
			low:  start,
			high: end,
		})

		if rand.Intn(10) > 7 {
			tree.Remove(avlint.Key{
				Val: start,
				Id:  i,
			})
		} else {
			array = append(array, &basicRange{
				low:  start,
				high: end,
			})
		}
	}

	const ntest = 10
	for i := 0; i < ntest; i++ {
		low := rand.Intn(max)
		r := &basicRange{
			low:  low,
			high: low + rand.Intn(200),
		}

		tree.Shift(10, 42)
		to := tree.Overlap(r)

		for j := range array {
			if array[j].Low() >= 10 {
				array[j].ShiftLow(42)
			}
			if array[j].High() >= 10 {
				array[j].ShiftHigh(42)
			}
		}
		ao := array.Overlap(r)

		sortIntervals(to)
		sortIntervals(ao)

		if len(to) != len(ao) {
			t.Errorf("len incorrect (%v): %v != %v (seed: %d)", r, len(to), len(ao), seed)
			continue
		}

		for j := range to {
			// fmt.Printf("%d (%v): %v != %v (seed: %d)\n", j, r, to[j], ao[j], seed)
			if to[j].Low() != ao[j].Low() || to[j].High() != ao[j].High() {
				t.Errorf("%d (%v): %v != %v (seed: %d)", j, r, to[j], ao[j], seed)
			}
		}
	}
}

func BenchmarkTreeInsert(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tree := &avlint.Tree{}
		const ninterval = 100000
		last := 0
		for i := 0; i < ninterval; i++ {
			start := last
			end := start + rand.Intn(20) + 1
			tree.Add(avlint.Key{
				Val: start,
				Id:  i,
			}, &basicRange{
				low:  start,
				high: end,
			})
			last = end
		}
	}
}

func BenchmarkArrayInsert(b *testing.B) {
	for n := 0; n < b.N; n++ {
		array := make(intervalArray, 0)
		const ninterval = 100000
		last := 0
		for i := 0; i < ninterval; i++ {
			start := last
			end := start + rand.Intn(20) + 1
			array = append(array, &basicRange{
				low:  start,
				high: end,
			})
			last = end
		}
	}
}

var dummy []avlint.Interval

func BenchmarkTreeOverlap(b *testing.B) {
	tree := &avlint.Tree{}
	const ninterval = 1000000
	last := 0
	for i := 0; i < ninterval; i++ {
		start := last
		end := start + rand.Intn(20) + 1
		tree.Add(avlint.Key{
			Val: start,
			Id:  i,
		}, &basicRange{
			low:  start,
			high: end,
		})
		last = end
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		low := rand.Intn(last)
		r := &basicRange{
			low:  low,
			high: low + rand.Intn(200),
		}
		dummy = tree.Overlap(r)
	}
}

func BenchmarkArrayOverlap(b *testing.B) {
	array := make(intervalArray, 0)
	const ninterval = 1000000
	last := 0
	for i := 0; i < ninterval; i++ {
		start := last
		end := start + rand.Intn(20) + 1
		array = append(array, &basicRange{
			low:  start,
			high: end,
		})
		last = end
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		low := rand.Intn(last)
		r := &basicRange{
			low:  low,
			high: low + rand.Intn(200),
		}
		dummy = array.Overlap(r)
	}
}
