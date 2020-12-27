// Package re provides functions for compiling 're' patterns (given as strings)
// into standard patterns.
package re

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/zyedidia/gpeg/capture"
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

// Peg grammar from Ford:
// # Hierarchical syntax
// Grammar    <- Spacing_ Definition+ EndOfFile_
// Definition <- Identifier LEFTARROW Expression
//
// Expression <- Sequence (SLASH Sequence)*
// Sequence   <- Prefix*
// Prefix     <- (AND / NOT)? Suffix
// Suffix     <- Primary (QUESTION / STAR / PLUS)?
// Primary    <- Identifier !LEFTARROW
//             / OPEN Expression CLOSE
//             / Literal / Class / DOT
//
// # Lexical syntax
// Identifier <- IdentStart IdentCont* Spacing_
// IdentStart <- [a-zA-Z_]
// IdentCont  <- IdentStart / [0-9]
//
// Literal    <- ['] (!['] Char)* ['] Spacing_
//             / ["] (!["] Char)* ["] Spacing_
// Class      <- '[' (!']' Range)* ']' Spacing_
// Range      <- Char '-' Char / Char
// Char       <- '\\' [\-nrt'"\[\]\\]
//             / '\\' [0-2][0-7][0-7]
//             / '\\' [0-7][0-7]?
//             / !'\\' .
//
// LEFTARROW  <- '<-' Spacing_
// SLASH      <- '/' Spacing_
// AND        <- '&' Spacing_
// NOT        <- '!' Spacing_
// QUESTION   <- '?' Spacing_
// STAR       <- '*' Spacing_
// PLUS       <- '+' Spacing_
// OPEN       <- '(' Spacing_
// CLOSE      <- ')' Spacing_
// DOT        <- '.' Spacing_
//
// Spacing_    <- (Space_ / Comment_)*
// Comment_    <- '#' (!EndOfLine_ .)* EndOfLine_
// Space_      <- ' ' / '\t' / EndOfLine_
// EndOfLine_  <- '\r\n' / '\n' / '\r'
// EndOfFile_  <- !.

var grammar = map[string]Pattern{
	"Grammar":    Concat(NonTerm("Spacing_"), Plus(NonTerm("Definition")), NonTerm("EndOfFile_")),
	"Definition": Concat(NonTerm("Identifier"), NonTerm("LEFTARROW"), NonTerm("Expression")),
	"Expression": Concat(NonTerm("Sequence"), Star(Concat(NonTerm("SLASH"), NonTerm("Sequence")))),
	"Sequence":   Star(NonTerm("Prefix")),
	"Prefix":     Concat(Optional(Or(NonTerm("AND"), NonTerm("NOT"))), NonTerm("Suffix")),
	"Suffix":     Concat(NonTerm("Primary"), Optional(Or(NonTerm("QUESTION"), NonTerm("STAR"), NonTerm("PLUS")))),
	"Primary": Or(Concat(NonTerm("Identifier"), Not(NonTerm("LEFTARROW"))),
		Concat(NonTerm("OPEN"), NonTerm("Expression"), NonTerm("CLOSE")),
		NonTerm("Literal"), NonTerm("Class"), NonTerm("DOT")),
	"Identifier": Concat(NonTerm("IdentStart"), Star(NonTerm("IdentCont")), NonTerm("Spacing_")),
	"IdentStart": Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')).Add(charset.New([]byte{'_'}))),
	"IdentCont":  Or(NonTerm("IdentStart"), Set(charset.Range('0', '9'))),
	"Literal": Or(Concat(Literal("'"), Star(Concat(Not(Literal("'")), NonTerm("Char"))), Literal("'"), NonTerm("Spacing_")),
		Concat(Literal("\""), Star(Concat(Not(Literal("\"")), NonTerm("Char"))), Literal("\""), NonTerm("Spacing_"))),
	"Class": Concat(Literal("["), Star(Concat(Not(Literal("]")), NonTerm("Range"))), Literal("]"), NonTerm("Spacing_")),
	"Range": Or(Concat(NonTerm("Char"), Literal("-"), NonTerm("Char")), NonTerm("Char")),
	"Char": Or(Concat(Literal("\\"), Set(charset.New([]byte{'n', 'r', 't', '\'', '"', '[', ']', '\\', '-'}))),
		Concat(Literal("\\"), Set(charset.Range('0', '2')), Set(charset.Range('0', '7')), Set(charset.Range('0', '7'))),
		Concat(Literal("\\"), Set(charset.Range('0', '7')), Optional(Set(charset.Range('0', '7')))),
		Concat(Not(Literal("\\")), Any(1))),

	"LEFTARROW": Concat(Literal("<-"), NonTerm("Spacing_")),
	"SLASH":     Concat(Literal("/"), NonTerm("Spacing_")),
	"AND":       Concat(Literal("&"), NonTerm("Spacing_")),
	"NOT":       Concat(Literal("!"), NonTerm("Spacing_")),
	"QUESTION":  Concat(Literal("?"), NonTerm("Spacing_")),
	"STAR":      Concat(Literal("*"), NonTerm("Spacing_")),
	"PLUS":      Concat(Literal("+"), NonTerm("Spacing_")),
	"OPEN":      Concat(Literal("("), NonTerm("Spacing_")),
	"CLOSE":     Concat(Literal(")"), NonTerm("Spacing_")),
	"DOT":       Concat(Literal("."), NonTerm("Spacing_")),

	"Spacing_":   Star(Or(NonTerm("Space_"), NonTerm("Comment_"))),
	"Comment_":   Concat(Literal("#"), Star(Concat(Not(NonTerm("EndOfLine_")), Any(1))), NonTerm("EndOfLine_")),
	"Space_":     Or(Literal(" "), Literal("\t"), NonTerm("EndOfLine_")),
	"EndOfLine_": Or(Literal("\r\n"), Literal("\n"), Literal("\r")),
	"EndOfFile_": Not(Any(1)),
}

var ids map[string]int16

var pegCode vm.VMCode

func init() {
	ids = make(map[string]int16)
	p := pattern.CapGrammar("Grammar", grammar, ids)
	pegCode = vm.Encode(MustCompile(p))
}

var special = map[byte]byte{
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'\'': '\'',
	'"':  '"',
	'[':  '[',
	']':  ']',
	'\\': '\\',
	'-':  '-',
}

func parseChar(char []byte) byte {
	switch char[0] {
	case '\\':
		if len(char) <= 1 {
			panic("Bad char")
		}
		for k, v := range special {
			if char[1] == k {
				return v
			}
		}

		i, err := strconv.ParseInt(string(char[1:]), 8, 8)
		if err != nil {
			panic("Bad char")
		}
		return byte(i)
	default:
		return char[0]
	}
}

func parseId(n *capture.Node, in *input.Input) string {
	ident := &bytes.Buffer{}
	for _, c := range n.Children {
		if c.Id != ids["IdentStart"] && c.Id != ids["IdentCont"] {
			continue
		}

		ident.Write(in.Slice(c.Start(), c.End()))
	}
	return ident.String()
}

func compileDef(n *capture.Node, in *input.Input) (string, pattern.Pattern) {
	id := n.Children[0]
	exp := n.Children[2]
	return parseId(id, in), compile(exp, in, nil)
}

func compileSet(n *capture.Node, in *input.Input) charset.Set {
	switch len(n.Children) {
	case 1:
		c := n.Children[0]
		return charset.New([]byte{parseChar(in.Slice(c.Start(), c.End()))})
	case 2:
		c1, c2 := n.Children[0], n.Children[1]
		return charset.Range(parseChar(in.Slice(c1.Start(), c1.End())), parseChar(in.Slice(c2.Start(), c2.End())))
	}
	return charset.Set{}
}

func compile(n *capture.Node, in *input.Input, capids map[string]int16) pattern.Pattern {
	var p pattern.Pattern
	switch n.Id {
	case ids["Grammar"]:
		nonterms := make(map[string]pattern.Pattern)
		var first string
		for _, c := range n.Children {
			if c.Id != ids["Definition"] {
				continue
			}

			k, v := compileDef(c, in)
			if first == "" {
				first = k
			}
			nonterms[k] = v
		}
		if capids != nil {
			p = pattern.CapGrammar(first, nonterms, capids)
		} else {
			p = pattern.Grammar(first, nonterms)
		}
	case ids["Expression"]:
		alternations := make([]pattern.Pattern, 0, len(n.Children))
		for _, c := range n.Children {
			if c.Id != ids["Sequence"] {
				continue
			}
			alternations = append(alternations, compile(c, in, nil))
		}
		p = pattern.Or(alternations...)
	case ids["Sequence"]:
		concats := make([]pattern.Pattern, 0, len(n.Children))
		for _, c := range n.Children {
			if c.Id != ids["Prefix"] {
				continue
			}
			concats = append(concats, compile(c, in, nil))
		}
		p = pattern.Concat(concats...)
	case ids["Prefix"]:
		c := n.Children[0]
		switch c.Id {
		case ids["AND"]:
			p = pattern.And(compile(n.Children[1], in, nil))
		case ids["NOT"]:
			p = pattern.Not(compile(n.Children[1], in, nil))
		default:
			p = compile(n.Children[0], in, nil)
		}
	case ids["Suffix"]:
		if len(n.Children) == 2 {
			c := n.Children[1]
			switch c.Id {
			case ids["QUESTION"]:
				p = pattern.Optional(compile(n.Children[0], in, nil))
			case ids["STAR"]:
				p = pattern.Star(compile(n.Children[0], in, nil))
			case ids["PLUS"]:
				p = pattern.Plus(compile(n.Children[0], in, nil))
			}
		} else {
			p = compile(n.Children[0], in, nil)
		}
	case ids["Primary"]:
		for _, c := range n.Children {
			switch c.Id {
			case ids["Identifier"], ids["Expression"], ids["Literal"], ids["Class"]:
				p = compile(c, in, nil)
			case ids["DOT"]:
				p = pattern.Any(1)
			default:
				continue
			}
			break
		}
	case ids["Literal"]:
		lit := &bytes.Buffer{}
		for _, c := range n.Children {
			if c.Id != ids["Char"] {
				continue
			}
			lit.WriteByte(parseChar(in.Slice(c.Start(), c.End())))
		}
		p = pattern.Literal(lit.String())
	case ids["Class"]:
		var set charset.Set
		for _, c := range n.Children {
			if c.Id != ids["Range"] {
				continue
			}

			set = set.Add(compileSet(c, in))
		}
		p = pattern.Set(set)
	case ids["Identifier"]:
		p = pattern.NonTerm(parseId(n, in))
	}
	return p
}

// CompileGrammar compiles the given pegexp grammar and captures every
// non-terminal. It also returns a map from non-terminal names to capture IDs.
func CompileGrammar(s string) (pattern.Pattern, map[string]int16, error) {
	in := input.StringReader(s)
	machine := vm.NewVM(in, pegCode)
	match, length, root, _ := machine.Exec(memo.NoneTable{})

	if !match {
		return nil, nil, errors.New("Not a valid PEG expression: failed at " + fmt.Sprintf("%v", length))
	}
	input := input.NewInput(in)

	capids := make(map[string]int16)
	return compile(root[0], input, capids), capids, nil
}

// CompilePatt attempts to compile the given pegexp pattern.
func CompilePatt(s string) (pattern.Pattern, error) {
	in := input.StringReader(s)
	machine := vm.NewVM(in, pegCode)
	match, length, root, _ := machine.Exec(memo.NoneTable{})

	if !match {
		return nil, errors.New("Not a valid PEG expression: failed at " + fmt.Sprintf("%v", length))
	}
	input := input.NewInput(in)

	return compile(root[0], input, nil), nil
}

// MustCompilePatt is the same as CompilePatt but panics on an error.
func MustCompilePatt(s string) pattern.Pattern {
	p, err := CompilePatt(s)
	if err != nil {
		panic(err)
	}
	return p
}
