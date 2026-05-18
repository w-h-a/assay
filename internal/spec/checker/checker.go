package checker

import (
	"fmt"
	"strings"

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
			c.define(d.Name, symbolType, "", d, d.Pos)
		case *ast.FuncDecl:
			c.define(d.Name, symbolFunc, "", d, d.Pos)
		case *ast.PredicateDecl:
			c.define(d.Name, symbolPredicate, "", d, d.Pos)
		case *ast.PropertyDecl:
			c.define(d.Name, symbolProperty, "", d, d.Pos)
		}
	}
}

// define registers a name in the current scope. If the name is already
// defined, it reports a duplicate error pointing at the new declaration
// and referencing the original
func (c *checker) define(name string, kind symbolKind, typeName string, decl ast.Decl, pos ast.Position) {
	if name == "" {
		return
	}

	if prev, exists := c.scope.symbols[name]; exists {
		c.addError(pos, "%s %q already declared at %s", prev.kind, name, prev.pos)
		return
	}

	c.scope.symbols[name] = symbol{kind: kind, typeName: typeName, decl: decl, pos: pos}
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
		c.define(p.Name, symbolVar, c.resolvedTypeName(p.Type), nil, p.Pos)
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
		c.define(v.Name, symbolVar, c.resolvedTypeName(v.Type), nil, v.Pos)
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
				c.define(s.Names[0], symbolVar, rhsType, nil, s.Pos)
			}
		} else {
			for _, name := range s.Names {
				if name != "_" {
					c.define(name, symbolVar, "", nil, s.Pos)
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
		operandType := c.inferType(e.Expr)
		if operandType != "" && !includesError(operandType) {
			c.addError(e.Pos, "'is %s' requires operand with error type, got %q", e.Target, operandType)
		}
		return "bool"
	case *ast.CallExpr:
		argTypes := make([]string, len(e.Args))
		for i, arg := range e.Args {
			argTypes[i] = c.inferType(arg)
		}

		sym, ok := c.scope.lookup(e.Func)
		if !ok {
			if retType, builtin := builtinFuncs[e.Func]; builtin {
				return retType
			}
			c.addError(e.Pos, "undefined function %q", e.Func)
			return ""
		}

		switch sym.kind {
		case symbolFunc:
			fd := sym.decl.(*ast.FuncDecl)
			if !c.checkCallArgs(e.Pos, e.Func, argTypes, fd.Params) {
				return ""
			}
			return c.funcReturnType(fd)
		case symbolPredicate:
			pd := sym.decl.(*ast.PredicateDecl)
			c.checkCallArgs(e.Pos, e.Func, argTypes, pd.Params)
			return "bool"
		default:
			c.addError(e.Pos, "may not call %s %q", sym.kind, e.Func)
			return ""
		}
	case *ast.FieldAccessExpr:
		objectType := c.inferType(e.Object)
		if objectType == "" {
			return ""
		}
		sym, ok := c.scope.lookup(objectType)
		if !ok || sym.kind != symbolType {
			c.addError(e.Pos, "field access on non-struct type %q", objectType)
			return ""
		}
		td := sym.decl.(*ast.TypeDecl)
		for _, f := range td.Fields {
			if f.Name == e.Field {
				return c.resolvedTypeName(f.Type)
			}
		}
		c.addError(e.Pos, "type %q has no field %q", objectType, e.Field)
		return ""
	case *ast.TupleExpr:
		parts := make([]string, len(e.Elements))
		allKnown := true
		for i, el := range e.Elements {
			resolved := c.inferType(el)
			if resolved == "" {
				allKnown = false
			}
			parts[i] = resolved
		}
		if !allKnown {
			return ""
		}
		return "(" + strings.Join(parts, ", ") + ")"
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

// checkCallArgs validates argument count and types against parameters.
// Returns false if the argument count is wrong.
func (c *checker) checkCallArgs(pos ast.Position, name string, argTypes []string, params []ast.Param) bool {
	if len(argTypes) != len(params) {
		c.addError(pos, "%q expects %d argument(s), got %d", name, len(params), len(argTypes))
		return false
	}
	for i, argType := range argTypes {
		paramType := c.resolvedTypeName(params[i].Type)
		if argType != "" && paramType != "" && argType != paramType {
			c.addError(pos, "argument %d to %q has type %q, expected %q", i+1, name, argType, paramType)
		}
	}
	return true
}

// funcReturnType computes the return type for a function declaration.
// Multiple returns are represented as a tuple type string.
func (c *checker) funcReturnType(fd *ast.FuncDecl) string {
	if len(fd.Returns) == 0 {
		return ""
	}
	if len(fd.Returns) == 1 {
		return c.resolvedTypeName(fd.Returns[0])
	}
	parts := make([]string, len(fd.Returns))
	for i, r := range fd.Returns {
		resolved := c.resolvedTypeName(r)
		if resolved == "" {
			return ""
		}
		parts[i] = resolved
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (c *checker) addError(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	})
}
