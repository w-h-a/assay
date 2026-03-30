package ast

import (
	"fmt"
	"strings"
)

// Decl is implemented by all declaration nodes in a spec.
type Decl interface {
	decNode()
}

// SpecDecl is the root node representing an entire spec.
type SpecDecl struct {
	Name         string
	Declarations []Decl
	Pos          Position
}

// TypeDecl declares a named type, optionally with record fields.
type TypeDecl struct {
	Name   string
	Fields []FieldDecl
	Pos    Position
}

func (*TypeDecl) decNode() {}

// FieldDecl is a field within a record TypeDecl.
type FieldDecl struct {
	Name string
	Type TypeExpr
	Pos  Position
}

// FuncDecl declares a function signature in a spec.
type FuncDecl struct {
	Name    string
	Params  []Param
	Returns []TypeExpr
	Pos     Position
}

func (*FuncDecl) decNode() {}

// Param is a named, typed parameter.
type Param struct {
	Name string
	Type TypeExpr
	Pos  Position
}

// TypeExpr represents a type expression in the spec language.
// Structural variants:
//   - Tuple: Name is empty, Elements is non-empty ((uint, error))
//   - Parameterized: Name is set, Params is non-empty (list[int], map[string, bytes])
//   - Named: Name is set, Params and Elements are nil (bool, int, MyType, error)
type TypeExpr struct {
	Name     string
	Params   []TypeExpr
	Elements []TypeExpr
	Pos      Position
}

func (t TypeExpr) String() string {
	switch {
	case len(t.Elements) > 0:
		elems := make([]string, len(t.Elements))
		for i, e := range t.Elements {
			elems[i] = e.String()
		}
		return "(" + strings.Join(elems, ", ") + ")"
	case len(t.Params) > 0:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = p.String()
		}
		return t.Name + "[" + strings.Join(params, ", ") + "]"
	default:
		return t.Name
	}
}

// Position tracks a location in source code.
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
