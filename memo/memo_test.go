package memo

import (
	"testing"

	"github.com/zyedidia/gpeg/input"
)

func TestEntry(t *testing.T) {
	tests := []struct {
		name        string
		match, exam int
	}{
		{"t1", 12, 13},
		{"t2", 12, 12},
		{"t3", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ent := NewEntry(tt.match, tt.exam, nil)
			if ent.MatchLength() != tt.match {
				t.Errorf("Incorrect match length %v should be %v", ent.MatchLength(), tt.match)
			}
			if ent.Examined() != tt.exam {
				t.Errorf("Incorrect examined length %v should be %v", ent.Examined(), tt.exam)
			}
		})
	}
}

func TestLRUTable(t *testing.T) {

	table := NewLRUTable(2)
	k1 := Key{1, input.Pos{0}}
	k2 := Key{2, input.Pos{10}}
	k3 := Key{8, input.Pos{17}}

	e1 := NewEntry(1, 1, nil)
	e2 := NewEntry(2, 2, nil)
	e3 := NewEntry(3, 3, nil)
	e4 := NewEntry(4, 4, nil)

	var ret Entry
	var ok bool

	table.Put(k1, e1)

	ret, ok = table.Get(k1)
	if !ok || ret.length != e1.length {
		t.Error("Incorrect entry for k1")
	}
	ret, ok = table.Get(k2)
	if ok {
		t.Error("Incorrect entry for k2")
	}

	table.Put(k2, e2)
	table.Put(k2, e3)
	ret, ok = table.Get(k2)
	if !ok || ret.length != e3.length {
		t.Error("Incorrect entry for k2")
	}
	ret, ok = table.Get(k1)
	if !ok || ret.length != e1.length {
		t.Error("Incorrect entry for k1")
	}

	table.Put(k3, e4)
	ret, ok = table.Get(k2)
	if ok {
		t.Error("Incorrect entry for k2")
	}
	ret, ok = table.Get(k3)
	if !ok || ret.length != e4.length {
		t.Error("Incorrect entry for k3")
	}
}
