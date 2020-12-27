package gpeg

import (
	"io/ioutil"
	"testing"

	"github.com/zyedidia/gpeg/re"
)

func TestRe(t *testing.T) {
	p := re.MustCompilePatt("ID <- [a-zA-Z][a-zA-Z0-9_]*")
	tests := []PatternTest{
		{"hello", 5},
		{"test_1", 6},
		{"_not_allowed", -1},
		{"123", -1},
	}
	check(p, tests, t)
}

func TestJson(t *testing.T) {
	peg, err := ioutil.ReadFile("grammars/json.peg")
	if err != nil {
		t.Error(err)
	}
	p := re.MustCompilePatt(string(peg))

	json, err := ioutil.ReadFile("testdata/test.json")
	if err != nil {
		t.Error(err)
	}

	tests := []PatternTest{
		{string(json), len(json)},
	}

	check(p, tests, t)
}

func TestJava(t *testing.T) {
	peg, err := ioutil.ReadFile("grammars/java.peg")
	if err != nil {
		t.Error(err)
	}
	p := re.MustCompilePatt(string(peg))

	java, err := ioutil.ReadFile("testdata/test.java")
	if err != nil {
		t.Error(err)
	}

	tests := []PatternTest{
		{string(java), len(java)},
	}

	check(p, tests, t)
}
