package pegexp

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/zyedidia/gpeg/ast"
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	. "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

// Peg grammar from Ford:
// # Hierarchical syntax
// Grammar    <- Spacing Definition+ EndOfFile
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
// Identifier <- IdentStart IdentCont* Spacing
// IdentStart <- [a-zA-Z_]
// IdentCont  <- IdentStart / [0-9]
//
// Literal    <- ['] (!['] Char)* ['] Spacing
//             / ["] (!["] Char)* ["] Spacing
// Class      <- '[' (!']' Range)* ']' Spacing
// Range      <- Char '-' Char / Char
// Char       <- '\\' [nrt'"\[\]\\]
//             / '\\' [0-2][0-7][0-7]
//             / '\\' [0-7][0-7]?
//             / !'\\' .
//
// LEFTARROW  <- '<-' Spacing
// SLASH      <- '/' Spacing
// AND        <- '&' Spacing
// NOT        <- '!' Spacing
// QUESTION   <- '?' Spacing
// STAR       <- '*' Spacing
// PLUS       <- '+' Spacing
// OPEN       <- '(' Spacing
// CLOSE      <- ')' Spacing
// DOT        <- '.' Spacing
//
// Spacing    <- (Space / Comment)*
// Comment    <- '#' (!EndOfLine .)* EndOfLine
// Space      <- ' ' / '\t' / EndOfLine
// EndOfLine  <- '\r\n' / '\n' / '\r'
// EndOfFile  <- !.

var grammar = map[string]Pattern{
	"Grammar":    Concat(NonTerm("Spacing"), Plus(NonTerm("Definition")), NonTerm("EndOfFile")),
	"Definition": Concat(NonTerm("Identifier"), NonTerm("LEFTARROW"), NonTerm("Expression")),
	"Expression": Concat(NonTerm("Sequence"), Star(Concat(NonTerm("SLASH"), NonTerm("Sequence")))),
	"Sequence":   Star(NonTerm("Prefix")),
	"Prefix":     Concat(Optional(Or(NonTerm("AND"), NonTerm("NOT"))), NonTerm("Suffix")),
	"Suffix":     Concat(NonTerm("Primary"), Optional(Or(NonTerm("QUESTION"), NonTerm("STAR"), NonTerm("PLUS")))),
	"Primary": Or(Concat(NonTerm("Identifier"), Not(NonTerm("LEFTARROW"))),
		Concat(NonTerm("OPEN"), NonTerm("Expression"), NonTerm("CLOSE")),
		NonTerm("Literal"), NonTerm("Class"), NonTerm("DOT")),
	"Identifier": Concat(NonTerm("IdentStart"), Star(NonTerm("IdentCont")), NonTerm("Spacing")),
	"IdentStart": Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z')).Add(charset.New([]byte{'_'}))),
	"IdentCont":  Or(NonTerm("IdentStart"), Set(charset.Range('0', '9'))),
	"Literal": Or(Concat(Literal("'"), Star(Concat(Not(Literal("'")), NonTerm("Char"))), Literal("'"), NonTerm("Spacing")),
		Concat(Literal("\""), Star(Concat(Not(Literal("\"")), NonTerm("Char"))), Literal("\""), NonTerm("Spacing"))),
	"Class": Concat(Literal("["), Star(Concat(Not(Literal("]")), NonTerm("Range"))), Literal("]"), NonTerm("Spacing")),
	"Range": Or(Concat(NonTerm("Char"), Literal("-"), NonTerm("Char")), NonTerm("Char")),
	"Char": Or(Concat(Literal("\\"), Set(charset.New([]byte{'n', 'r', 't', '[', ']', '\\'}))),
		Concat(Literal("\\"), Set(charset.Range('0', '2')), Set(charset.Range('0', '7')), Set(charset.Range('0', '7'))),
		Concat(Literal("\\"), Set(charset.Range('0', '7')), Optional(Set(charset.Range('0', '7')))),
		Concat(Not(Literal("\\")), Any(1))),

	"LEFTARROW": Concat(Literal("<-"), NonTerm("Spacing")),
	"SLASH":     Concat(Literal("/"), NonTerm("Spacing")),
	"AND":       Concat(Literal("&"), NonTerm("Spacing")),
	"NOT":       Concat(Literal("!"), NonTerm("Spacing")),
	"QUESTION":  Concat(Literal("?"), NonTerm("Spacing")),
	"STAR":      Concat(Literal("*"), NonTerm("Spacing")),
	"PLUS":      Concat(Literal("+"), NonTerm("Spacing")),
	"OPEN":      Concat(Literal("("), NonTerm("Spacing")),
	"CLOSE":     Concat(Literal(")"), NonTerm("Spacing")),
	"DOT":       Concat(Literal("."), NonTerm("Spacing")),

	"Spacing":   Star(Or(NonTerm("Space"), NonTerm("Comment"))),
	"Comment":   Concat(Literal("#"), Star(Concat(Not(NonTerm("EndOfLine")), Any(1))), NonTerm("EndOfLine")),
	"Space":     Or(Literal(" "), Literal("\t"), NonTerm("EndOfLine")),
	"EndOfLine": Or(Literal("\r\n"), Literal("\n"), Literal("\r")),
	"EndOfFile": Not(Any(1)),
}

var ids map[string]int16

var pegCode vm.VMCode

func init() {
	ids = make(map[string]int16)
	p := pattern.CapGrammar("Grammar", grammar, ids)
	pegCode = vm.Encode(p)
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

func parseId(n *ast.Node, in input.Reader) string {
	ident := &bytes.Buffer{}
	for _, c := range n.Children {
		if c.Id != ids["IdentStart"] && c.Id != ids["IdentCont"] {
			continue
		}

		ident.Write(in.Slice(c.Start, c.End))
	}
	return ident.String()
}

func compileDef(n *ast.Node, in input.Reader) (string, pattern.Pattern) {
	id := n.Children[0]
	exp := n.Children[2]
	return parseId(id, in), compile(exp, in)
}

func compileSet(n *ast.Node, in input.Reader) charset.Set {
	switch len(n.Children) {
	case 1:
		c := n.Children[0]
		return charset.New([]byte{parseChar(in.Slice(c.Start, c.End))})
	case 2:
		c1, c2 := n.Children[0], n.Children[1]
		return charset.Range(parseChar(in.Slice(c1.Start, c1.End)), parseChar(in.Slice(c2.Start, c2.End)))
	}
	return charset.Set{}
}

func compile(n *ast.Node, in input.Reader) pattern.Pattern {
	p := make(pattern.Pattern, 0)
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
		p = pattern.Grammar(first, nonterms)
	case ids["Expression"]:
		alternations := make([]pattern.Pattern, 0, len(n.Children))
		for _, c := range n.Children {
			if c.Id != ids["Sequence"] {
				continue
			}
			alternations = append(alternations, compile(c, in))
		}
		p = pattern.Or(alternations...)
	case ids["Sequence"]:
		concats := make([]pattern.Pattern, 0, len(n.Children))
		for _, c := range n.Children {
			if c.Id != ids["Prefix"] {
				continue
			}
			concats = append(concats, compile(c, in))
		}
		p = pattern.Concat(concats...)
	case ids["Prefix"]:
		c := n.Children[0]
		switch c.Id {
		case ids["AND"]:
			p = pattern.And(compile(n.Children[1], in))
		case ids["NOT"]:
			p = pattern.Not(compile(n.Children[1], in))
		default:
			p = compile(n.Children[0], in)
		}
	case ids["Suffix"]:
		if len(n.Children) == 2 {
			c := n.Children[1]
			switch c.Id {
			case ids["QUESTION"]:
				p = pattern.Optional(compile(n.Children[0], in))
			case ids["STAR"]:
				p = pattern.Star(compile(n.Children[0], in))
			case ids["PLUS"]:
				p = pattern.Plus(compile(n.Children[0], in))
			}
		} else {
			p = compile(n.Children[0], in)
		}
	case ids["Primary"]:
		for _, c := range n.Children {
			switch c.Id {
			case ids["Identifier"], ids["Expression"], ids["Literal"], ids["Class"]:
				p = compile(c, in)
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
			lit.WriteByte(parseChar(in.Slice(c.Start, c.End)))
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

func CompilePatt(s string) (pattern.Pattern, error) {
	in := input.StringReader(s)
	machine := vm.NewVM(in, pegCode)
	match, _, caps := machine.Exec(memo.NoneTable{})

	if !match {
		return nil, errors.New("Not a valid PEG expression")
	}

	root := machine.CaptureAST(caps, pegCode)
	return compile(root[0], in), nil
}

func MustCompilePatt(s string) pattern.Pattern {
	p, err := CompilePatt(s)
	if err != nil {
		panic(err)
	}
	return p
}
