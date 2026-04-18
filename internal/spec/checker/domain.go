package checker

import "github.com/w-h-a/assay/internal/spec/ast"

type symbolKind string

const (
	symbolType      symbolKind = "type"
	symbolFunc      symbolKind = "func"
	symbolPredicate symbolKind = "predicate"
	symbolProperty  symbolKind = "property"
)

type symbol struct {
	kind symbolKind
	pos  ast.Position
}
