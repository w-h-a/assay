package lexer

import "fmt"

// TokenKind identifies the type of a lexical token.
type TokenKind int

func (k TokenKind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return fmt.Sprintf("TokenKind(%d)", k)
}

const (
	// Special
	ILLEGAL TokenKind = iota
	EOF

	// Identifiers and literals
	IDENT
	INT_LIT
	FLOAT_LIT
	STRING_LIT

	// Keywords — spec language
	SPEC
	TYPE
	FUNC
	PREDICATE
	PROPERTY
	FORALL
	WHERE
	LET
	REQUIRE
	ASSERT
	IS
	OK
	ERROR
	AND
	OR
	NOT
	IN
	TRUE
	FALSE

	// Keywords — binding language
	BIND
	TARGET
	PACKAGE

	// Keywords — type names
	BOOL
	INT
	UINT
	FLOAT
	STRING
	BYTES
	LIST
	SET
	MAP
	OPTION

	// Operators
	EQ      // ==
	NEQ     // !=
	LT      // <
	LTE     // <=
	GT      // >
	GTE     // >=
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %
	ASSIGN  // =
	DOTDOT  // ..
	ARROW   // ->

	// Delimiters
	LPAREN     // (
	RPAREN     // )
	LBRACE     // {
	RBRACE     // }
	LBRACKET   // [
	RBRACKET   // ]
	COMMA      // ,
	COLON      // :
	DOT        // .
	UNDERSCORE // _
)

var kindNames = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	IDENT:      "IDENT",
	INT_LIT:    "INT_LIT",
	FLOAT_LIT:  "FLOAT_LIT",
	STRING_LIT: "STRING_LIT",
	SPEC:       "spec",
	TYPE:       "type",
	FUNC:       "func",
	PREDICATE:  "predicate",
	PROPERTY:   "property",
	FORALL:     "forall",
	WHERE:      "where",
	LET:        "let",
	REQUIRE:    "require",
	ASSERT:     "assert",
	IS:         "is",
	OK:         "ok",
	ERROR:      "error",
	AND:        "and",
	OR:         "or",
	NOT:        "not",
	IN:         "in",
	TRUE:       "true",
	FALSE:      "false",
	BIND:       "bind",
	TARGET:     "target",
	PACKAGE:    "package",
	BOOL:       "bool",
	INT:        "int",
	UINT:       "uint",
	FLOAT:      "float",
	STRING:     "string",
	BYTES:      "bytes",
	LIST:       "list",
	SET:        "set",
	MAP:        "map",
	OPTION:     "option",
	EQ:         "==",
	NEQ:        "!=",
	LT:         "<",
	LTE:        "<=",
	GT:         ">",
	GTE:        ">=",
	PLUS:       "+",
	MINUS:      "-",
	STAR:       "*",
	SLASH:      "/",
	PERCENT:    "%",
	ASSIGN:     "=",
	DOTDOT:     "..",
	ARROW:      "->",
	LPAREN:     "(",
	RPAREN:     ")",
	LBRACE:     "{",
	RBRACE:     "}",
	LBRACKET:   "[",
	RBRACKET:   "]",
	COMMA:      ",",
	COLON:      ":",
	DOT:        ".",
	UNDERSCORE: "_",
}

var keywords = map[string]TokenKind{
	"spec":      SPEC,
	"type":      TYPE,
	"func":      FUNC,
	"predicate": PREDICATE,
	"property":  PROPERTY,
	"forall":    FORALL,
	"where":     WHERE,
	"let":       LET,
	"require":   REQUIRE,
	"assert":    ASSERT,
	"is":        IS,
	"ok":        OK,
	"error":     ERROR,
	"and":       AND,
	"or":        OR,
	"not":       NOT,
	"in":        IN,
	"true":      TRUE,
	"false":     FALSE,
	"bind":      BIND,
	"target":    TARGET,
	"package":   PACKAGE,
	"bool":      BOOL,
	"int":       INT,
	"uint":      UINT,
	"float":     FLOAT,
	"string":    STRING,
	"bytes":     BYTES,
	"list":      LIST,
	"set":       SET,
	"map":       MAP,
	"option":    OPTION,
}

// LookupKeyword returns the keyword TokenKind for word if it is a reserved
// keyword, or IDENT if it is a user-defined name.
func LookupKeyword(word string) TokenKind {
	if kind, ok := keywords[word]; ok {
		return kind
	}
	return IDENT
}

// Token represents a lexical token with its kind, literal text, and source position.
type Token struct {
	Kind    TokenKind
	Literal string
	Pos     Position
}

// Position represents a location in a source file.
type Position struct {
	File   string
	Line   int
	Column int
}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}
