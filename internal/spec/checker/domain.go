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

var builtinTypes = map[string]bool{
	"bool":   true,
	"int":    true,
	"uint":   true,
	"float":  true,
	"string": true,
	"bytes":  true,
	"error":  true,
	"list":   true,
	"set":    true,
	"map":    true,
	"option": true,
}
