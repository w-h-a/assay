package lexer

import "fmt"

// TokenKind identifies the type of a lexical token in the binding language.
type TokenKind int

func (k TokenKind) String() string {
	if k >= 0 && int(k) < len(kindNames) {
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
	STRING_LIT

	// Keywords
	BIND
	TARGET
	PACKAGE
	TYPE
	FUNC

	// Operators
	ASSIGN // =

	// Delimiters
	LBRACE // {
	RBRACE // }
	DOT    // .
)

var kindNames = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	IDENT:      "IDENT",
	STRING_LIT: "STRING_LIT",
	BIND:       "bind",
	TARGET:     "target",
	PACKAGE:    "package",
	TYPE:       "type",
	FUNC:       "func",
	ASSIGN:     "=",
	LBRACE:     "{",
	RBRACE:     "}",
	DOT:        ".",
}

var keywords = map[string]TokenKind{
	"bind":    BIND,
	"target":  TARGET,
	"package": PACKAGE,
	"type":    TYPE,
	"func":    FUNC,
}

// LookupKeyword returns the keyword TokenKind for word if it is a reserved
// keyword in the binding language, or IDENT if it is a user-defined name.
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
