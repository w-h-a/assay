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

// builtinTypes maps builtin type names to their arity.
var builtinTypes = map[string]int{
	"bool":   0,
	"int":    0,
	"uint":   0,
	"float":  0,
	"string": 0,
	"bytes":  0,
	"error":  0,
	"list":   1,
	"set":    1,
	"map":    2,
	"option": 1,
}
