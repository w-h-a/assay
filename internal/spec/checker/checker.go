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
		scope: &scope{symbols: map[string]symbol{}},
	}

	c.registerDeclarations(spec.Declarations)

	c.resolveDeclarations(spec.Declarations)

	return &ast.ValidatedSpec{Spec: spec}, c.errors
}

// checker walks a parsed spec AST to validate declarations.
// Errors are collected rather than halting, so a single check pass
// can report multiple problems.
type checker struct {
	scope  *scope
	errors []Error
}

func (c *checker) pushScope() {
	c.scope = &scope{symbols: map[string]symbol{}, parent: c.scope}
}

func (c *checker) popScope() {
	c.scope = c.scope.parent
}

// registerDeclarations walks all declarations and registers their
// names in the environment, reporting duplicates.
func (c *checker) registerDeclarations(decls []ast.Decl) {
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			c.define(d.Name, symbolType, "", d.Pos)
		case *ast.FuncDecl:
			c.define(d.Name, symbolFunc, "", d.Pos)
		case *ast.PredicateDecl:
			c.define(d.Name, symbolPredicate, "", d.Pos)
		case *ast.PropertyDecl:
			c.define(d.Name, symbolProperty, "", d.Pos)
		}
	}
}

// define registers a name in the current scope. If the name is already
// defined, it reports a duplicate error pointing at the new declaration
// and referencing the original
func (c *checker) define(name string, kind symbolKind, typeName string, pos ast.Position) {
	if name == "" {
		return
	}

	if prev, exists := c.scope.symbols[name]; exists {
		c.addError(pos, "%s %q already declared at %s", prev.kind, name, prev.pos)
		return
	}

	c.scope.symbols[name] = symbol{kind: kind, typeName: typeName, pos: pos}
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
		case *ast.PredicateDecl:
			for _, p := range d.Params {
				c.resolveTypeExpr(p.Type)
			}
			c.checkPredicateBody(d)
		case *ast.PropertyDecl:
			for _, v := range d.Forall.Vars {
				c.resolveTypeExpr(v.Type)
			}
			c.checkPropertyBody(d)
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
	s, exists := c.scope.lookup(name)
	return exists && s.kind == symbolType
}

// checkPredicateBody validates that a predicate body does not call
// spec-declared functions.
//   - spec-declared functions may not be called
//   - the body expression must be boolean
//   - predicate params are available in body scope
func (c *checker) checkPredicateBody(decl *ast.PredicateDecl) {
	if decl.Body == nil {
		return
	}

	c.pushScope()
	defer c.popScope()

	for _, p := range decl.Params {
		c.define(p.Name, symbolVar, c.resolvedTypeName(p.Type), p.Pos)
	}

	c.rejectSpecFuncCalls(decl.Body, "predicate body")

	bodyType := c.inferType(decl.Body)
	if bodyType != "" && bodyType != "bool" {
		c.addError(decl.Pos, "predicate body must be a boolean expression, got %q", bodyType)
	}
}

// checkPropertyBody validates a property body:
//   - forall vars are available in scope
//   - where clause must be boolean (if present)
//   - require conditions must be boolean
//   - terminal assertions must be boolean
//   - let bindings introduce names into scope
func (c *checker) checkPropertyBody(decl *ast.PropertyDecl) {
	c.pushScope()
	defer c.popScope()

	for _, v := range decl.Forall.Vars {
		c.define(v.Name, symbolVar, c.resolvedTypeName(v.Type), v.Pos)
	}

	if decl.Where != nil {
		c.rejectSpecFuncCalls(decl.Where.Condition, "where clause")
		whereType := c.inferType(decl.Where.Condition)
		if whereType != "" && whereType != "bool" {
			c.addError(decl.Where.Pos, "where clause must be a boolean expression, got %q", whereType)
		}
	}

	for _, stmt := range decl.Body {
		c.checkStmt(stmt)
	}
}

func (c *checker) checkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.LetBinding:
		rhsType := c.inferType(s.Expr)
		if len(s.Names) == 1 {
			if s.Names[0] != "_" {
				c.define(s.Names[0], symbolVar, rhsType, s.Pos)
			}
		} else {
			for _, name := range s.Names {
				if name != "_" {
					c.define(name, symbolVar, "", s.Pos)
				}
			}
		}
	case *ast.RequireStmt:
		reqType := c.inferType(s.Expr)
		if reqType != "" && reqType != "bool" {
			c.addError(s.Pos, "require condition must be a boolean expression, got %q", reqType)
		}
	case *ast.AssertExpr:
		assertType := c.inferType(s.Expr)
		if assertType != "" && assertType != "bool" {
			c.addError(s.Pos, "property assertion must be a boolean expression, got %q", assertType)
		}
	}
}

// resolvedTypeName returns the type name for use in expression inference.
// If the type expression refers to an unknown type, it returns "" to
// prevent cascading errors from downstream inference.
func (c *checker) resolvedTypeName(te ast.TypeExpr) string {
	switch {
	case len(te.Elements) > 0:
		return te.String()
	case len(te.Params) > 0:
		if te.Name != "" && !c.isKnownType(te.Name) {
			return ""
		}
		return te.String()
	default:
		if te.Name == "" || !c.isKnownType(te.Name) {
			return ""
		}
		return te.Name
	}
}

// rejectSpecFuncCalls walks an expression tree and reports an error
// for each CallExpr that references a spec-declared function.
func (c *checker) rejectSpecFuncCalls(expr ast.Expr, context string) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		if s, ok := c.scope.lookup(e.Func); ok && s.kind == symbolFunc {
			c.addError(e.Pos, "cannot call function %q in %s", e.Func, context)
		}
		for _, arg := range e.Args {
			c.rejectSpecFuncCalls(arg, context)
		}
	case *ast.BinaryExpr:
		c.rejectSpecFuncCalls(e.Left, context)
		c.rejectSpecFuncCalls(e.Right, context)
	case *ast.UnaryExpr:
		c.rejectSpecFuncCalls(e.Operand, context)
	case *ast.IsExpr:
		c.rejectSpecFuncCalls(e.Expr, context)
	case *ast.FieldAccessExpr:
		c.rejectSpecFuncCalls(e.Object, context)
	case *ast.TupleExpr:
		for _, el := range e.Elements {
			c.rejectSpecFuncCalls(el, context)
		}
	case *ast.IdentExpr, *ast.LiteralExpr:
		// leaf nodes — nothing to check
	}
}

// inferType walks an expression and returns its inferred type.
// Unknown types are returned as "". Callers skip constraints
// that depend on an unknown operand to avoid cascading errors.
func (c *checker) inferType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return literalType(e.Kind)
	case *ast.IdentExpr:
		sym, ok := c.scope.lookup(e.Name)
		if !ok {
			c.addError(e.Pos, "undefined identifier %q", e.Name)
			return ""
		}
		return sym.typeName
	case *ast.BinaryExpr:
		return c.inferBinaryType(e)
	case *ast.UnaryExpr:
		return c.inferUnaryType(e)
	case *ast.IsExpr:
		c.inferType(e.Expr)
		return "bool"
	case *ast.CallExpr:
		for _, arg := range e.Args {
			c.inferType(arg)
		}
		return ""
	case *ast.FieldAccessExpr:
		c.inferType(e.Object)
		return ""
	case *ast.TupleExpr:
		for _, el := range e.Elements {
			c.inferType(el)
		}
		return ""
	default:
		return ""
	}
}

func (c *checker) inferBinaryType(e *ast.BinaryExpr) string {
	left := c.inferType(e.Left)
	right := c.inferType(e.Right)

	switch e.Op {
	case "+", "-", "*", "/", "%":
		if left != "" && !isNumeric(left) {
			c.addError(e.Pos, "operator %q requires numeric operands, got %q", e.Op, left)
			return ""
		}
		if right != "" && !isNumeric(right) {
			c.addError(e.Pos, "operator %q requires numeric operands, got %q", e.Op, right)
			return ""
		}
		if left != "" && right != "" && left != right {
			c.addError(e.Pos, "operator %q requires matching numeric types, got %q and %q", e.Op, left, right)
			return ""
		}
		if left != "" {
			return left
		}
		return right
	case "==", "!=", "<", ">", "<=", ">=":
		if left != "" && right != "" && left != right {
			if !(isNumeric(left) && isNumeric(right)) {
				c.addError(e.Pos, "operator %q requires matching types, got %q and %q", e.Op, left, right)
			}
		}
		return "bool"
	case "and", "or":
		if left != "" && left != "bool" {
			c.addError(e.Pos, "operator %q requires bool operands, got %q", e.Op, left)
		}
		if right != "" && right != "bool" {
			c.addError(e.Pos, "operator %q requires bool operands, got %q", e.Op, right)
		}
		return "bool"
	default:
		return ""
	}
}

func (c *checker) inferUnaryType(e *ast.UnaryExpr) string {
	operand := c.inferType(e.Operand)

	switch e.Op {
	case "not":
		if operand != "" && operand != "bool" {
			c.addError(e.Pos, "operator %q requires bool operand, got %q", e.Op, operand)
		}
		return "bool"
	case "-":
		if operand != "" && !isNumeric(operand) {
			c.addError(e.Pos, "unary %q requires numeric operand, got %q", e.Op, operand)
			return ""
		}
		return operand
	default:
		return ""
	}
}

func (c *checker) addError(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	})
}
