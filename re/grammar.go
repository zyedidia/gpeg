package re

import (
	"github.com/zyedidia/gpeg/charset"
	. "github.com/zyedidia/gpeg/pattern"
)

// Pattern    <- Spacing_ (Grammar / Expression) EndOfFile_
// Grammar    <- Definition+
// Definition <- Identifier LEFTARROW Expression
//
// Expression <- Sequence (SLASH Sequence)*
// Sequence   <- Prefix*
// Prefix     <- (AND / NOT)? Suffix
// Suffix     <- Primary (QUESTION / STAR / PLUS)?
// Primary    <- Identifier !LEFTARROW
// 			/ '(' Expression ')'
// 			/ Literal / Class
// 			/ BRACEO Expression BRACEC
// 			/ BRACEPO Expression BRACEPC
// 			/ DOT
//
// Identifier <- IdentStart IdentCont* Spacing_
// IdentStart <- [a-zA-Z_]
// IdentCont  <- IdentStart / [0-9]
//
// Literal    <- ['] (!['] Char)* ['] Spacing_
// 			/ ["] (!["] Char)* ["] Spacing_
// Class      <- '[' CARAT? (!']' Range)* ']' Spacing_
// Range      <- Char '-' Char / Char
// Char       <- '\\' [nrt'"\[\]\\\-]
// 			/ '\\' [0-2][0-7][0-7]
// 			/ '\\' [0-7][0-7]?
// 			/ !'\\' .
//
// AND        <- '&' Spacing_
// NOT        <- '!' Spacing_
// QUESTION   <- '?' Spacing_
// STAR       <- '*' Spacing_
// PLUS       <- '+' Spacing_
// DOT        <- '.' Spacing_
// CARAT      <- '^' Spacing_
// BRACEO     <- '{' Spacing_
// BRACEC     <- '}' Spacing_
// BRACEPO    <- '{+' Spacing_
// BRACEPC    <- '+}' Spacing_
// LEFTARROW  <- '<-' Spacing_
// OPEN       <- '(' Spacing_
// CLOSE      <- ')' Spacing_
// SLASH      <- '/' Spacing_
//
// Spacing_   <- (Space_ / Comment_)*
// Comment_   <- '#' (!EndOfLine_ .)* EndOfLine_
// Space_     <- ' ' / '\t' / EndOfLine_
// EndOfLine_ <- '\r\n' / '\n' / '\r'
// EndOfFile_ <- !.

const (
	idPattern = iota
	idGrammar
	idDefinition
	idExpression
	idSequence
	idPrefix
	idSuffix
	idPrimary
	idLiteral
	idRange
	idClass
	idIdentifier
	idIdentStart
	idIdentCont
	idChar
	idAND
	idNOT
	idQUESTION
	idSTAR
	idPLUS
	idDOT
	idCARAT
	idOPEN
	idBRACEO
	idBRACEPO
)

var grammar = map[string]Pattern{
	"Pattern": CapId(Concat(
		NonTerm("Spacing"),
		Or(
			NonTerm("Grammar"),
			NonTerm("Expression"),
		),
		NonTerm("EndOfFile"),
	), idPattern),
	"Grammar": CapId(Plus(NonTerm("Definition")), idGrammar),
	"Definition": CapId(Concat(
		NonTerm("Identifier"),
		NonTerm("LEFTARROW"),
		NonTerm("Expression"),
	), idDefinition),

	"Expression": CapId(Concat(
		NonTerm("Sequence"),
		Star(Concat(
			NonTerm("SLASH"),
			NonTerm("Sequence"),
		)),
	), idExpression),
	"Sequence": CapId(Star(NonTerm("Prefix")), idSequence),
	"Prefix": CapId(Concat(
		Optional(Or(
			NonTerm("AND"),
			NonTerm("NOT"),
		)),
		NonTerm("Suffix"),
	), idPrefix),
	"Suffix": CapId(Concat(
		NonTerm("Primary"),
		Optional(Or(
			NonTerm("QUESTION"),
			NonTerm("STAR"),
			NonTerm("PLUS"),
		)),
	), idSuffix),
	"Primary": CapId(Or(
		Concat(
			NonTerm("Identifier"),
			Not(NonTerm("LEFTARROW")),
		),
		Concat(
			NonTerm("OPEN"),
			NonTerm("Expression"),
			NonTerm("CLOSE"),
		),
		NonTerm("Literal"),
		NonTerm("Class"),
		Concat(
			NonTerm("BRACEO"),
			NonTerm("Expression"),
			NonTerm("BRACEC"),
		),
		Concat(
			NonTerm("BRACEPO"),
			NonTerm("Expression"),
			NonTerm("BRACEPC"),
		),
		NonTerm("DOT"),
	), idPrimary),

	"Identifier": CapId(Concat(
		NonTerm("IdentStart"),
		Star(NonTerm("IdentCont")),
		NonTerm("Spacing"),
	), idIdentifier),
	"IdentStart": CapId(
		Set(charset.Range('a', 'z').
			Add(charset.Range('A', 'Z')).
			Add(charset.New([]byte{'_'})),
		), idIdentStart),
	"IdentCont": CapId(Or(
		NonTerm("IdentStart"),
		Set(charset.Range('0', '9')),
	), idIdentCont),

	"Literal": CapId(Or(
		Concat(
			Literal("'"),
			Star(Concat(
				Not(Literal("'")),
				NonTerm("Char"),
			)),
			Literal("'"),
			NonTerm("Spacing"),
		),
		Concat(
			Literal("\""),
			Star(Concat(
				Not(Literal("\"")),
				NonTerm("Char"),
			)),
			Literal("\""),
			NonTerm("Spacing"),
		),
	), idLiteral),
	"Class": CapId(Concat(
		Literal("["),
		Optional(NonTerm("CARAT")),
		Star(Concat(
			Not(Literal("]")),
			NonTerm("Range"),
		)),
		Literal("]"),
		NonTerm("Spacing"),
	), idClass),
	"Range": CapId(Or(
		Concat(
			NonTerm("Char"),
			Literal("-"),
			NonTerm("Char"),
		),
		NonTerm("Char"),
	), idRange),
	"Char": CapId(Or(
		Concat(
			Literal("\\"),
			Set(charset.New([]byte{'n', 'r', 't', '\'', '"', '[', ']', '\\', '-'})),
		),
		Concat(
			Literal("\\"),
			Set(charset.Range('0', '2')),
			Set(charset.Range('0', '7')),
			Set(charset.Range('0', '7')),
		),
		Concat(
			Literal("\\"),
			Set(charset.Range('0', '7')),
			Optional(Set(charset.Range('0', '7'))),
		),
		Concat(
			Not(Literal("\\")),
			Any(1),
		),
	), idChar),

	"AND": CapId(Concat(
		Literal("&"),
		NonTerm("Spacing"),
	), idAND),
	"NOT": CapId(Concat(
		Literal("!"),
		NonTerm("Spacing"),
	), idNOT),
	"QUESTION": CapId(Concat(
		Literal("?"),
		NonTerm("Spacing"),
	), idQUESTION),
	"STAR": CapId(Concat(
		Literal("*"),
		NonTerm("Spacing"),
	), idSTAR),
	"PLUS": CapId(Concat(
		Literal("+"),
		NonTerm("Spacing"),
	), idPLUS),
	"DOT": CapId(Concat(
		Literal("."),
		NonTerm("Spacing"),
	), idDOT),
	"CARAT": CapId(Concat(
		Literal("^"),
		NonTerm("Spacing"),
	), idCARAT),
	"OPEN": CapId(Concat(
		Literal("("),
		NonTerm("Spacing"),
	), idOPEN),
	"CLOSE": Concat(
		Literal(")"),
		NonTerm("Spacing"),
	),
	"BRACEO": CapId(Concat(
		Literal("{"),
		NonTerm("Spacing"),
	), idBRACEO),
	"BRACEC": Concat(
		Literal("}"),
		NonTerm("Spacing"),
	),
	"BRACEPO": CapId(Concat(
		Literal("{+"),
		NonTerm("Spacing"),
	), idBRACEPO),
	"BRACEPC": Concat(
		Literal("+}"),
		NonTerm("Spacing"),
	),
	"SLASH": Concat(
		Literal("/"),
		NonTerm("Spacing"),
	),
	"LEFTARROW": Concat(
		Literal("<-"),
		NonTerm("Spacing"),
	),

	"Spacing": Star(Or(
		NonTerm("Space"),
		NonTerm("Comment"),
	)),
	"Comment": Concat(
		Literal("#"),
		Star(Concat(
			Not(NonTerm("EndOfLine")),
			Any(1),
		)),
		NonTerm("EndOfLine"),
	),
	"Space": Or(
		Set(charset.New([]byte{' ', '\t'})),
		NonTerm("EndOfLine"),
	),
	"EndOfLine": Or(
		Literal("\r\n"),
		Literal("\n"),
		Literal("\r"),
	),
	"EndOfFile": Not(Any(1)),
}
