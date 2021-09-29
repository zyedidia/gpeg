package rxconv_test

import (
	"regexp/syntax"
	"strings"
	"testing"

	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/rxconv"
	"github.com/zyedidia/gpeg/vm"
)

type PatternTest struct {
	in    string
	match int
}

func check(p Pattern, tests []PatternTest, t *testing.T) {
	code := vm.Encode(MustCompile(p))
	for _, tt := range tests {
		name := tt.in[:min(10, len(tt.in))]
		t.Run(name, func(t *testing.T) {
			match, off, _, _ := code.Exec(strings.NewReader(tt.in), memo.NoneTable{})
			if tt.match == -1 && match || tt.match != -1 && !match || tt.match != -1 && tt.match != off {
				t.Errorf("%s: got: (%t, %d), but expected (%d)\n", tt.in, match, off, tt.match)
			}
		})
	}
}

func TestSimple(t *testing.T) {
	peg, err := rxconv.FromRegexp("(a|ab)c", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}

	tests := []PatternTest{
		{"abc", 3},
		{"ac", 2},
		{"ab", -1},
	}
	check(peg, tests, t)
}

func TestStar(t *testing.T) {
	peg, err := rxconv.FromRegexp("(ba|a)*a", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}

	tests := []PatternTest{
		{"abaabaa", 7},
	}

	check(peg, tests, t)
}

func TestMultiOr(t *testing.T) {
	peg, err := rxconv.FromRegexp("aa|bb|dd|ff", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}
	tests := []PatternTest{
		{"aa", 2},
		{"bb", 2},
		{"af", -1},
		{"ff", 2},
	}

	check(peg, tests, t)
}

func TestCharClass(t *testing.T) {
	peg, err := rxconv.FromRegexp("[a-z0-9]+", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}
	tests := []PatternTest{
		{"", -1},
		{"hello123", 8},
		{"foo", 3},
		{"123", 3},
		{"_&_", -1},
	}
	check(peg, tests, t)
}

func TestEmptyOp(t *testing.T) {
	peg, err := rxconv.FromRegexp("^foo", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}
	tests := []PatternTest{
		{"foohello", 3},
		{" foo ", -1},
	}
	check(peg, tests, t)

	peg, err = rxconv.FromRegexp("\\bfoo\\b", syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}
	tests = []PatternTest{
		{"foohello", -1},
		{" foo ", 4},
	}
	check(peg, tests, t)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
