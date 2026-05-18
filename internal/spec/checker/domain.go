package checker

import "github.com/w-h-a/assay/internal/spec/ast"

type symbolKind string

const (
	symbolType      symbolKind = "type"
	symbolFunc      symbolKind = "func"
	symbolPredicate symbolKind = "predicate"
	symbolProperty  symbolKind = "property"
	symbolVar       symbolKind = "var"
)

type symbol struct {
	kind     symbolKind
	typeName string
	decl     ast.Decl
	pos      ast.Position
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

// builtinFuncs maps builtin function names to their return types.
var builtinFuncs = map[string]string{
	"len": "int",
}

// scope is a lexical scope mapping names to symbols.
// Scopes form a chain: lookups check the current scope first,
// then walk up to parent scopes.
type scope struct {
	symbols map[string]symbol
	parent  *scope
}

func (s *scope) lookup(name string) (symbol, bool) {
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}
	if s.parent != nil {
		return s.parent.lookup(name)
	}
	return symbol{}, false
}
