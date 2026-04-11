package parser

import (
	"github.com/w-h-a/assay/internal/spec/ast"
	"github.com/w-h-a/assay/internal/spec/lexer"
)

// precedence represents the binding power of a binary operator
// in the Pratt expression parser.
type precedence int

const (
	precNone           precedence = iota
	precOr                        // or
	precAnd                       // and
	precComparison                // == != < > <= >=
	precAddition                  // + -
	precMultiplication            // * / %
)

// binaryPrec returns the precedence of a binary operator token,
// or precNone if the token is not a binary operator.
func binaryPrec(kind lexer.TokenKind) precedence {
	switch kind {
	case lexer.OR:
		return precOr
	case lexer.AND:
		return precAnd
	case lexer.EQ, lexer.NEQ, lexer.LT, lexer.GT, lexer.LTE, lexer.GTE:
		return precComparison
	case lexer.PLUS, lexer.MINUS:
		return precAddition
	case lexer.STAR, lexer.SLASH, lexer.PERCENT:
		return precMultiplication
	default:
		return precNone
	}
}

// isTypeName reports whether a token kind can appear as a type name.
func isTypeName(kind lexer.TokenKind) bool {
	switch kind {
	case lexer.IDENT,
		lexer.BOOL, lexer.INT, lexer.UINT, lexer.FLOAT,
		lexer.STRING, lexer.BYTES,
		lexer.LIST, lexer.SET, lexer.MAP, lexer.OPTION,
		lexer.ERROR:
		return true
	default:
		return false
	}
}

// astPos converts a lexer token position to an AST position.
func astPos(tok lexer.Token) ast.Position {
	return ast.Position{
		File:   tok.Pos.File,
		Line:   tok.Pos.Line,
		Column: tok.Pos.Column,
	}
}
