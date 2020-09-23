package gpeg

import (
	"testing"

	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/isa"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

type PatternTest struct {
	in    string
	match int
}

func check(p Pattern, tests []PatternTest, t *testing.T) {
	code := vm.Encode(p)
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var bytes input.ByteReader = []byte(tt.in)
			machine := vm.NewVM(bytes, 0)
			nchars := machine.Exec(code)
			if tt.match != nchars {
				t.Errorf("%v returned %v", string(bytes), nchars)
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

	p = Concat(Plus(Set(isa.NewCharset([]byte{'0', '1'}))), Star(Set(isa.NewCharset([]byte{'a', 'b', 'c'}))))
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
	p := Plus(Set(isa.CharsetRange('0', '9')))
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
	// S <- <B> / (![()] .)*
	// B <- '(' <S> ')'
	S := Or(NonTerm("B"), Plus(Concat(Not(Set(isa.NewCharset([]byte{'(', ')'}))), Any(1))))
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
	p := Plus(Or(Set(isa.CharsetRange('a', 'z')), Set(isa.CharsetRange('A', 'Z'))))
	tests := []PatternTest{
		{"Hello", 5},
		{"123", -1},
		{"Hello1", 5},
	}
	check(p, tests, t)
}

func TestArithmeticGrammar(t *testing.T) {
	// grammar:
	// Expr   <- <Factor> ([+-] <Factor>)*
	// Factor <- <Term> ([*/] <Term>)*
	// Term   <- <Number> / '(' <Expr> ')'
	// Number <- [0-9]+
	p := Grammar("Expr", map[string]Pattern{
		"Expr":   Concat(NonTerm("Factor"), Star(Concat(Set(isa.NewCharset([]byte{'+', '-'})), NonTerm("Factor")))),
		"Factor": Concat(NonTerm("Term"), Star(Concat(Set(isa.NewCharset([]byte{'*', '/'})), NonTerm("Term")))),
		"Term":   Or(NonTerm("Number"), Concat(Concat(Literal("("), NonTerm("Expr")), Literal(")"))),
		"Number": Plus(Set(isa.CharsetRange('0', '9'))),
	})
	tests := []PatternTest{
		{"13+(22-15)", 10},
		{"24*5+3", 6},
		{"word 5*3", -1},
		{"10*(43", 2},
	}
	check(p, tests, t)
}

var arith input.ByteReader = []byte(`21 + 27 + 18 / 8 * 28 * 97 - (22 - 5 + 44) * 1 - 100 / 64 + 42 * 56 + 32 - 71 - 38 + 80 + 58 - 30 / 20 * 79 + 22 * (97 / 59) / 47 - 22 - 35 / 29 + (2 + 83) + 96 / 13 / 84 + 95 / (61 * 19 / 57 + 35) * 90 - 65 - 3 - 25 + 12 - 88 + 80 * 39 * 89 * 100 * 32 + 88 + 17 + 51 + 48 - (41 / 21) + 78 / 18 + 59 * 78 * 85 / 90 + 5 * 22 + 77 - 92 - 6 / 26 * 77 / 9 - 19 + 1 - 65 / 81 - 56`)

var match int

func BenchmarkArithmeticGrammar(b *testing.B) {
	p := Grammar("Expr", map[string]Pattern{
		"Expr":   Concat(NonTerm("Factor"), Star(Concat(Set(isa.NewCharset([]byte{'+', '-'})), NonTerm("Factor")))),
		"Factor": Concat(NonTerm("Term"), Star(Concat(Set(isa.NewCharset([]byte{'*', '/'})), NonTerm("Term")))),
		"Term":   Or(NonTerm("Number"), Concat(Concat(Literal("("), NonTerm("Expr")), Literal(")"))),
		"Number": Plus(Set(isa.CharsetRange('0', '9'))),
	})
	code := vm.Encode(p)
	machine := vm.NewVM(arith, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		match = machine.Exec(code)
		machine.Reset(0)
	}
}

var words input.ByteReader = []byte(`able about account acid across act addition adjustment
advertisement after again against agreement almost among
attempt attention attraction authority automatic awake
baby back bad bag balance ball band base basin basket bath be
beautiful because bed bee before behaviour belief bell
bent berry between bird birth bit bite bitter black blade blood
carriage cart cat cause certain chain chalk chance
change cheap cheese chemical chest chief chin church circle clean clear
clock cloth cloud coal coat cold collar colour comb
come comfort committee common company comparison competition complete
complex condition connection conscious control cook copper copy
cord cork cotton cough country cover cow crack credit crime
delicate dependent design desire destruction detail development
different digestion direction dirty discovery discussion disease
last late laugh law lead leaf learning leather left letter level
library lift light like limit line linen lip liquid
morning mother motion mountain mouth move much muscle music nail
name narrow nation natural near necessary neck need needle
private probable process produce profit property prose protest public
pull pump punishment purpose push put quality question
seem selection self send sense separate serious servant shade shake
shame sharp sheep shelf ship shirt shock shoe short
square stage stamp star start statement station steam steel stem step
stick sticky stiff still stitch stocking stomach stone
stop store story straight strange street stretch strong structure
substance such sudden sugar suggestion summer sun support surprise
very vessel view violent voice waiting walk wall war warm wash waste
watch water wave wax way weather week weight well west
wet wheel when where while whip whistle white who why wide will wind
window wine wing winter wire wise with woman wood wool word
work worm wound writing wrong year yellow yesterday young`)

func BenchmarkWordSearch(b *testing.B) {
	p := Literal("story")
	p = Grammar("S", map[string]Pattern{
		"S": Or(p, Concat(Any(1), NonTerm("S"))),
	})
	code := vm.Encode(p)
	machine := vm.NewVM(arith, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		match = machine.Exec(code)
		machine.Reset(0)
	}
}
