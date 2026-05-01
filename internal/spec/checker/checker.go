package checker

import (
	"fmt"

	"github.com/w-h-a/assay/internal/spec/ast"
)

// Error represents a type-checking error with source position.
type Error struct {
	Message string
	Pos     ast.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Check type-checks a parsed spec. It returns a validated spec and
// any errors found. The validated spec is always returned, even when
// errors are present, so callers can report multiple problems.
func Check(spec *ast.SpecDecl) (*ast.ValidatedSpec, []Error) {
	c := &checker{
		env: map[string]symbol{},
	}

	c.registerDeclarations(spec.Declarations)

	c.resolveDeclarations(spec.Declarations)

	return &ast.ValidatedSpec{Spec: spec}, c.errors
}

// checker walks a parsed spec AST to validate declarations.
// Errors are collected rather than halting, so a single check pass
// can report multiple problems.
type checker struct {
	env    map[string]symbol
	errors []Error
}

// registerDeclarations walks all declarations and registers their
// names in the environment, reporting duplicates.
func (c *checker) registerDeclarations(decls []ast.Decl) {
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			c.define(d.Name, symbolType, d.Pos)
		case *ast.FuncDecl:
			c.define(d.Name, symbolFunc, d.Pos)
		case *ast.PredicateDecl:
			c.define(d.Name, symbolPredicate, d.Pos)
		case *ast.PropertyDecl:
			c.define(d.Name, symbolProperty, d.Pos)
		}
	}
}

// define registers a name in the environment. If the name is already
// defined, it reports a duplicate error pointing at the new declaration
// and referencing the original
func (c *checker) define(name string, kind symbolKind, pos ast.Position) {
	if name == "" {
		return
	}

	if prev, exists := c.env[name]; exists {
		c.addError(pos, "%s %q already declared at %s", prev.kind, name, prev.pos)
		return
	}

	c.env[name] = symbol{kind: kind, pos: pos}
}

// resolveDeclarations walks declarations and checks that all type
// references resolve to known types
func (c *checker) resolveDeclarations(decls []ast.Decl) {
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			for _, f := range d.Fields {
				c.resolveTypeExpr(f.Type)
			}
		case *ast.FuncDecl:
			for _, p := range d.Params {
				c.resolveTypeExpr(p.Type)
			}
			for _, r := range d.Returns {
				c.resolveTypeExpr(r)
			}
		}
	}
}

// resolveTypeExpr checks that a type expression refers to known types.
// It recurses into parameterized types and tuple elements.
func (c *checker) resolveTypeExpr(te ast.TypeExpr) {
	switch {
	case len(te.Elements) > 0:
		for _, e := range te.Elements {
			c.resolveTypeExpr(e)
		}
	case len(te.Params) > 0:
		if te.Name != "" && !c.isKnownType(te.Name) {
			c.addError(te.Pos, "undefined type %q", te.Name)
		}
		if arity, ok := builtinTypes[te.Name]; ok {
			if arity == 0 {
				c.addError(te.Pos, "type %q does not accept type parameters", te.Name)
			} else if len(te.Params) != arity {
				c.addError(te.Pos, "type %q expects %d type parameter(s), got %d", te.Name, arity, len(te.Params))
			}
		} else if te.Name != "" && c.isKnownType(te.Name) {
			c.addError(te.Pos, "type %q does not accept type parameters", te.Name)
		}
		for _, p := range te.Params {
			c.resolveTypeExpr(p)
		}
	default:
		if te.Name == "" {
			return
		}
		if !c.isKnownType(te.Name) {
			c.addError(te.Pos, "undefined type %q", te.Name)
			return
		}
		if arity, ok := builtinTypes[te.Name]; ok && arity > 0 {
			c.addError(te.Pos, "type %q expects %d type parameter(s), got 0", te.Name, arity)
		}
	}
}

func (c *checker) isKnownType(name string) bool {
	if _, ok := builtinTypes[name]; ok {
		return true
	}
	s, exists := c.env[name]
	return exists && s.kind == symbolType
}

func (c *checker) addError(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	})
}
