package interval

import (
	"math/rand"
	"testing"
)

func randrange(max int) (int, int) {
	low := rand.Intn(max)
	high := low + rand.Intn(1000)
	if low == high {
		high = low + 1
	}
	return low, high
}

func randint(min, max int) int {
	return rand.Intn(max-min) + min
}

func checkParents(n *node, t *testing.T) {
	if n == nil {
		return
	}
	if n.left != nil && n.left.parent != n {
		t.Fatalf("Incorrect parent n: %p, n.left.parent: %p", n, n.left.parent)
	}
	if n.right != nil && n.right.parent != n {
		t.Fatalf("Incorrect parent n: %p, n.right.parent: %p", n, n.right.parent)
	}
	checkParents(n.left, t)
	checkParents(n.right, t)
}

func TestTree(t *testing.T) {
	it := &Tree{}
	ia := &Array{}

	const (
		opAdd = iota
		opFind
		opRemoveAndShift
		opPos

		nops     = 3000
		maxidx   = 100000
		maxid    = 10
		maxshamt = 50
	)

	var pt, pa Pos
	var length int
	var haspt bool

	for i := 0; i < nops; i++ {
		op := rand.Intn(4)
		switch op {
		case opAdd:
			id := rand.Intn(maxid)
			low, high := randrange(maxidx)
			pt = it.Add(id, low, high, i)
			pa = ia.Add(id, low, high, i)
			length = high - low
			haspt = true
		case opFind:
			id := rand.Intn(maxid)
			pos := rand.Intn(maxidx)

			vt := it.FindLargest(id, pos)
			va := ia.FindLargest(id, pos)

			if vt == nil && va == nil {
				continue
			}

			if vt == nil && va != nil || va == nil && vt != nil {
				t.Fatalf("Find (%d, %d): %v != %v", id, pos, vt, va)
			}

			if vt.(int) != va.(int) {
				t.Fatalf("Find (%d, %d): %d != %d", id, pos, vt.(int), va.(int))
			}
		case opRemoveAndShift:
			low, high := randrange(maxidx)
			amt := randint(-maxshamt, maxshamt)

			if haspt {
				ptpos := pt.Pos()
				if Overlaps(Interval{
					Low:  low,
					High: high,
				}, Interval{
					Low:  ptpos,
					High: ptpos + length,
				}) {
					haspt = false
				}
			}

			it.RemoveAndShift(low, high, amt)
			ia.RemoveAndShift(low, high, amt)
		case opPos:
			if haspt && pt.Pos() != pa.Pos() {
				t.Fatalf("%d != %d", pt.Pos(), pa.Pos())
			}
		}
		checkParents(it.root, t)
	}
}
