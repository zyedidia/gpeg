package gpeg

import (
	"testing"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

type PatternTest struct {
	in    string
	match int
}

func check(p Pattern, tests []PatternTest, t *testing.T) {
	code := vm.Encode(MustCompile(p))
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var bytes input.ByteReader = []byte(tt.in)
			machine := vm.NewVM(bytes, code)
			match, off, _ := machine.Exec(memo.NoneTable{})
			if tt.match == -1 && match || tt.match != -1 && tt.match != int(off) {
				t.Errorf("%v returned %v, %v", string(bytes), match, off)
			}
		})
	}
}

func TestConcat(t *testing.T) {
	p := Concat(Literal("ana"), Literal("hi"))

	tests := []PatternTest{
		{"ana", -1},
		{"hi", -1},
		{"anahi", 5},
		{"anah", -1},
	}

	check(p, tests, t)
}

func TestOr(t *testing.T) {
	p := Or(Literal("ana"), Literal("hi"))

	tests := []PatternTest{
		{"ana", 3},
		{"hi", 2},
		{"an", -1},
		{"anahi", 3},
	}

	check(p, tests, t)
}

func TestRepeat(t *testing.T) {
	p := Star(Literal("ana"))
	tests := []PatternTest{
		{"", 0},
		{"ana", 3},
		{"anaanaana", 9},
		{"hiana", 0},
		{"anaanaan", 6},
		{"an", 0},
	}
	check(p, tests, t)

	p = Plus(Literal("hi"))
	tests = []PatternTest{
		{"", -1},
		{"hi", 2},
		{"hihihi", 6},
		{"hihiana", 4},
		{"h", -1},
	}
	check(p, tests, t)

	p = Concat(Plus(Set(charset.New([]byte{'0', '1'}))), Star(Set(charset.New([]byte{'a', 'b', 'c'}))))
	tests = []PatternTest{
		{"01", 2},
		{"01abaabbc", 9},
		{"abc", -1},
		{"5a", -1},
		{"1z", 1},
	}
	check(p, tests, t)
}

func TestPredicate(t *testing.T) {
	p := Not(Literal("ana"))
	tests := []PatternTest{
		{"ana", -1},
		{"hi", 0},
		{"an", 0},
	}
	check(p, tests, t)

	p1 := Not(Not(Literal("ana")))
	p2 := And(Literal("ana"))
	tests = []PatternTest{
		{"ana", 0},
		{"hi", -1},
		{"an", -1},
	}
	check(p1, tests, t)
	check(p2, tests, t)
}

func TestAny(t *testing.T) {
	p := Concat(Any(5), Literal("ana"))
	tests := []PatternTest{
		{"helloana", 8},
		{"hiana", -1},
		{"anaanana", 8},
	}
	check(p, tests, t)
}

func TestOptional(t *testing.T) {
	p := Concat(Literal("ana"), Optional(Literal("hello")))
	tests := []PatternTest{
		{"ana", 3},
		{"anahe", 3},
		{"hello", -1},
		{"anahello", 8},
	}
	check(p, tests, t)
}

func TestSet(t *testing.T) {
	p := Plus(Set(charset.Range('0', '9')))
	tests := []PatternTest{
		{"hi", -1},
		{"1002", 4},
		{"10.02", 2},
		{"9", 1},
	}
	check(p, tests, t)
}

func TestGrammar(t *testing.T) {
	// grammar:
	// S <- <B> / (![()] .)+
	// B <- '(' <S> ')'
	S := Or(NonTerm("B"), Plus(Concat(Not(Set(charset.New([]byte{'(', ')'}))), Any(1))))
	B := Concat(Concat(Literal("("), NonTerm("S")), Literal(")"))

	p := Grammar("S", map[string]Pattern{
		"S": S,
		"B": B,
	})
	tests := []PatternTest{
		{"(hello)", 7},
		{"(hello", -1},
		{"((inside))", 10},
		{"((inside)", -1},
	}
	check(p, tests, t)
}

func TestTailCall(t *testing.T) {
	p := Grammar("X", map[string]Pattern{
		"X": Or(Literal("ana"), Concat(Any(1), NonTerm("X"))),
	})
	tests := []PatternTest{
		{"asdf", -1},
		{"ana hello", 3},
		{"hello ana", 9},
		{"anaana", 3},
	}
	check(p, tests, t)
}

func TestUnionSet(t *testing.T) {
	p := Plus(Or(Set(charset.Range('a', 'z')), Set(charset.Range('A', 'Z'))))
	tests := []PatternTest{
		{"Hello", 5},
		{"123", -1},
		{"Hello1", 5},
	}
	check(p, tests, t)
}

// func TestSearch(t *testing.T) {
// 	p := Search(Literal("ana"))
// 	tests := []PatternTest{
// 		{"hello ana hello", 9},
// 		{"hello", -1},
// 		{"hello ana ana ana", 9},
// 	}
// 	check(p, tests, t)
//
// 	// search for last occurrence
// 	p = Plus(Search(Literal("ana")))
// 	tests = []PatternTest{
// 		{"hello ana hello", 9},
// 		{"hello", -1},
// 		{"hello ana ana ana hello", 17},
// 	}
// 	check(p, tests, t)
// }

// func TestCapture(t *testing.T) {
// 	wordChar := charset.Range('A', 'Z').Add(charset.Range('a', 'z'))
// 	p := Star(Concat(Star(Set(wordChar.Complement())), Cap(Plus(Set(wordChar)))))
// }

func TestArithmeticGrammar(t *testing.T) {
	// grammar:
	// Expr   <- <Factor> ([+-] <Factor>)*
	// Factor <- <Term> ([*/] <Term>)*
	// Term   <- <Number> / '(' <Expr> ')'
	// Number <- [0-9]+
	p := Grammar("Expr", map[string]Pattern{
		"Expr":   Concat(NonTerm("Factor"), Star(Concat(Set(charset.New([]byte{'+', '-'})), NonTerm("Factor")))),
		"Factor": Concat(NonTerm("Term"), Star(Concat(Set(charset.New([]byte{'*', '/'})), NonTerm("Term")))),
		"Term":   Or(NonTerm("Number"), Concat(Concat(Literal("("), NonTerm("Expr")), Literal(")"))),
		"Number": Plus(Set(charset.Range('0', '9'))),
	})
	tests := []PatternTest{
		{"13+(22-15)", 10},
		{"24*5+3", 6},
		{"word 5*3", -1},
		{"10*(43", 2},
	}
	check(p, tests, t)
}

// **************
// * Benchmarks *
// **************

// var match bool
// var bible input.ByteReader
// var machine *vm.VM
//
// func TestMain(m *testing.M) {
// 	var err error
// 	bible, err = ioutil.ReadFile("testdata/bible.txt")
// 	if err != nil {
// 		fmt.Println("Warning:", err)
// 	}
// 	os.Exit(m.Run())
// }
//
// func BenchmarkBibleSearchFirstEartt(b *testing.B) {
// 	code := vm.Encode(Search(Literal("eartt")))
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }
//
// func BenchmarkBibleSearchFirstAbram(b *testing.B) {
// 	abram := Concat(Plus(Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')))), Literal(" Abram"))
// 	code := vm.Encode(Search(abram))
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }
//
// func BenchmarkBibleSearchLastAbram(b *testing.B) {
// 	abram := Concat(Plus(Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')))), Literal(" Abram"))
// 	code := vm.Encode(Star(Search(abram)))
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }
//
// func BenchmarkBibleSearchLastTubalcain(b *testing.B) {
// 	code := vm.Encode(Star(Search(Literal("Tubalcain"))))
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }
//
// func BenchmarkBibleOmegaPattern(b *testing.B) {
// 	omega := Concat(Star(Concat(Not(Literal("Omega")), Any(1))), Literal("Omega"))
// 	code := vm.Encode(omega)
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }
//
// func BenchmarkBibleOmegaGrammar(b *testing.B) {
// 	omega := Grammar("S", map[string]Pattern{
// 		"S": Concat(Star(Concat(Not(NonTerm("P")), Any(1))), NonTerm("P")),
// 		"P": Literal("Omega"),
// 	})
// 	code := vm.Encode(omega)
// 	machine := vm.NewVM(bible, code)
// 	machine.Reset()
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		match, _, _ = machine.Exec(memo.NoneTable{})
// 		machine.Reset()
// 		machine.SeekTo(0)
// 	}
// }