package parser

import (
	"github.com/w-h-a/assay/internal/spec/ast"
	"github.com/w-h-a/assay/internal/spec/lexer"
)

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
