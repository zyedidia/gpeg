package pegexp

import (
	"github.com/zyedidia/gpeg/charset"
	. "github.com/zyedidia/gpeg/pattern"
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
