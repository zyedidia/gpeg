package memo

import "testing"

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
			ent := NewEntry(tt.match, tt.exam)
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
	k1 := Key{1, 0}
	k2 := Key{2, 10}
	k3 := Key{8, 17}

	e1 := NewEntry(1, 1)
	e2 := NewEntry(2, 2)
	e3 := NewEntry(3, 3)
	e4 := NewEntry(4, 4)

	var ret Entry
	var ok bool

	table.Put(k1, e1)

	ret, ok = table.Get(k1)
	if !ok || ret != e1 {
		t.Error("Incorrect entry for k1")
	}
	ret, ok = table.Get(k2)
	if ok {
		t.Error("Incorrect entry for k2")
	}

	table.Put(k2, e2)
	table.Put(k2, e3)
	ret, ok = table.Get(k2)
	if !ok || ret != e3 {
		t.Error("Incorrect entry for k2")
	}
	ret, ok = table.Get(k1)
	if !ok || ret != e1 {
		t.Error("Incorrect entry for k1")
	}

	table.Put(k3, e4)
	ret, ok = table.Get(k2)
	if ok {
		t.Error("Incorrect entry for k2")
	}
	ret, ok = table.Get(k3)
	if !ok || ret != e4 {
		t.Error("Incorrect entry for k3")
	}
}