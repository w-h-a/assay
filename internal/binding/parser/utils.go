package parser

import (
	"github.com/w-h-a/assay/internal/binding/ast"
	"github.com/w-h-a/assay/internal/binding/lexer"
)

// astPos converts a lexer token position to an AST position.
func astPos(tok lexer.Token) ast.Position {
	return ast.Position{
		File:   tok.Pos.File,
		Line:   tok.Pos.Line,
		Column: tok.Pos.Column,
	}
}
