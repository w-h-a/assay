package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/w-h-a/assay/internal/spec/ast"
)

func TestParseEmptySpec(t *testing.T) {
	// arrange
	source := `spec "empty" {}`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Equal(t, "empty", spec.Name)
	require.Empty(t, spec.Declarations)
	require.Equal(t, 1, spec.Pos.Line)
	require.Equal(t, 1, spec.Pos.Column)
}

func TestParseStructType(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Point {
                        x: int,
                        y: int
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 1)

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Equal(t, "Point", td.Name)
	require.Len(t, td.Fields, 2)

	require.Equal(t, "x", td.Fields[0].Name)
	require.Equal(t, "int", td.Fields[0].Type.Name)

	require.Equal(t, "y", td.Fields[1].Name)
	require.Equal(t, "int", td.Fields[1].Type.Name)
}

func TestParseNonStructType(t *testing.T) {
	// arrange
	source := `spec "test" { type Log }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Equal(t, "test", spec.Name)
	require.Len(t, spec.Declarations, 1)

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Equal(t, "Log", td.Name)
	require.Empty(t, td.Fields)
}

func TestParseMultipleTypes(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Log
                type Entry {
                        offset: uint,
                        data: bytes
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 2)

	logType := spec.Declarations[0].(*ast.TypeDecl)
	require.Equal(t, "Log", logType.Name)
	require.Empty(t, logType.Fields)

	entryType := spec.Declarations[1].(*ast.TypeDecl)
	require.Equal(t, "Entry", entryType.Name)
	require.Len(t, entryType.Fields, 2)
	require.Equal(t, "offset", entryType.Fields[0].Name)
	require.Equal(t, "uint", entryType.Fields[0].Type.Name)
	require.Equal(t, "data", entryType.Fields[1].Name)
	require.Equal(t, "bytes", entryType.Fields[1].Type.Name)
}

func TestParseParameterizedType(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Store {
                        items: list[int],
                        labels: map[string, bytes]
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Len(t, td.Fields, 2)

	items := td.Fields[0].Type
	require.Equal(t, "list", items.Name)
	require.Len(t, items.Params, 1)
	require.Equal(t, "int", items.Params[0].Name)

	labels := td.Fields[1].Type
	require.Equal(t, "map", labels.Name)
	require.Len(t, labels.Params, 2)
	require.Equal(t, "string", labels.Params[0].Name)
	require.Equal(t, "bytes", labels.Params[1].Name)
}

func TestParseTupleType(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Wrapper {
                        result: (uint, error)
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Len(t, td.Fields, 1)

	result := td.Fields[0].Type
	require.Empty(t, result.Name)
	require.Len(t, result.Elements, 2)
	require.Equal(t, "uint", result.Elements[0].Name)
	require.Equal(t, "error", result.Elements[1].Name)
}

func TestParseNestedParameterizedType(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Container {
                        items: option[list[int]]
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	items := td.Fields[0].Type
	require.Equal(t, "option", items.Name)
	require.Len(t, items.Params, 1)
	require.Equal(t, "list", items.Params[0].Name)
	require.Len(t, items.Params[0].Params, 1)
	require.Equal(t, "int", items.Params[0].Params[0].Name)
}

func TestParseTrailingComma(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Point {
                        x: int,
                        y: int,
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Len(t, td.Fields, 2)
}

func TestParseErrorMissingSpecName(t *testing.T) {
	// arrange
	source := `spec {}`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.Len(t, errs, 1)
	require.Equal(t, 1, errs[0].Pos.Line)
}

func TestParseErrorMissingOpenBrace(t *testing.T) {
	// arrange
	source := `spec "test" type Foo`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Len(t, errs, 1)
	require.Equal(t, 1, errs[0].Pos.Line)
	require.Equal(t, "test", spec.Name)
}

func TestParseErrorMissingTypeName(t *testing.T) {
	// arrange
	source := `spec "test" { type }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
	require.Equal(t, 1, errs[0].Pos.Line)
}

func TestParseErrorMissingCloseBrace(t *testing.T) {
	// arrange
	source := `spec "test" { type Foo`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorMalformedField(t *testing.T) {
	// arrange
	source := `spec "test" {                                                               
                type Foo {
                        123: int
                }
        }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorMissingComma(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Point {
                        x: int
                        y: int
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "expected comma")

	td := spec.Declarations[0].(*ast.TypeDecl)
	require.Len(t, td.Fields, 2)
}

func TestParseTrailingCommaInTypeParams(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Store {
                        items: list[int,]
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	items := td.Fields[0].Type
	require.Equal(t, "list", items.Name)
	require.Len(t, items.Params, 1)
	require.Equal(t, "int", items.Params[0].Name)
}

func TestParseTrailingCommaInTuple(t *testing.T) {
	// arrange
	source := `spec "test" {
                type Wrapper {
                        result: (uint, error,)
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	td := spec.Declarations[0].(*ast.TypeDecl)
	result := td.Fields[0].Type
	require.Empty(t, result.Name)
	require.Len(t, result.Elements, 2)
	require.Equal(t, "uint", result.Elements[0].Name)
	require.Equal(t, "error", result.Elements[1].Name)
}

func TestParseFuncSingleReturn(t *testing.T) {
	// arrange
	source := `spec "test" {
                func append(log: Log, data: bytes) -> uint
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 1)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "append", fd.Name)
	require.Len(t, fd.Params, 2)

	require.Equal(t, "log", fd.Params[0].Name)
	require.Equal(t, "Log", fd.Params[0].Type.Name)
	require.Equal(t, "data", fd.Params[1].Name)
	require.Equal(t, "bytes", fd.Params[1].Type.Name)

	require.Len(t, fd.Returns, 1)
	require.Equal(t, "uint", fd.Returns[0].Name)
}

func TestParseFuncTupleReturn(t *testing.T) {
	// arrange
	source := `spec "test" {
                func append(log: Log, data: bytes) -> (uint, error)
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "append", fd.Name)
	require.Len(t, fd.Params, 2)
	require.Len(t, fd.Returns, 2)
	require.Equal(t, "uint", fd.Returns[0].Name)
	require.Equal(t, "error", fd.Returns[1].Name)
}

func TestParseFuncNoReturn(t *testing.T) {
	// arrange
	source := `spec "test" {
                func clear(log: Log)
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "clear", fd.Name)
	require.Len(t, fd.Params, 1)
	require.Equal(t, "log", fd.Params[0].Name)
	require.Equal(t, "Log", fd.Params[0].Type.Name)
	require.Empty(t, fd.Returns)
}

func TestParseFuncParameterizedTypes(t *testing.T) {
	// arrange
	source := `spec "test" {
                func get(store: map[string, bytes], key: string) -> option[bytes]
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "get", fd.Name)

	require.Len(t, fd.Params, 2)
	require.Equal(t, "store", fd.Params[0].Name)
	require.Equal(t, "map", fd.Params[0].Type.Name)
	require.Len(t, fd.Params[0].Type.Params, 2)
	require.Equal(t, "string", fd.Params[0].Type.Params[0].Name)
	require.Equal(t, "bytes", fd.Params[0].Type.Params[1].Name)

	require.Equal(t, "key", fd.Params[1].Name)
	require.Equal(t, "string", fd.Params[1].Type.Name)

	require.Len(t, fd.Returns, 1)
	require.Equal(t, "option", fd.Returns[0].Name)
	require.Len(t, fd.Returns[0].Params, 1)
	require.Equal(t, "bytes", fd.Returns[0].Params[0].Name)
}

func TestParseFuncNoParams(t *testing.T) {
	// arrange
	source := `spec "test" {
                func now() -> uint
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "now", fd.Name)
	require.Empty(t, fd.Params)
	require.Len(t, fd.Returns, 1)
	require.Equal(t, "uint", fd.Returns[0].Name)
}

func TestParseFuncErrorMissingName(t *testing.T) {
	// arrange
	source := `spec "test" { func (x: int) -> int }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseFuncErrorMissingParen(t *testing.T) {
	// arrange
	source := `spec "test" { func append log: Log -> uint }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseFuncErrorMalformedParam(t *testing.T) {
	// arrange
	source := `spec "test" { func append(123: int) -> uint }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseFuncTrailingCommaInParams(t *testing.T) {
	// arrange
	source := `spec "test" {                                                                                                    
                  func append(log: Log, data: bytes,) -> uint                                                                       
          }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	fd := spec.Declarations[0].(*ast.FuncDecl)
	require.Equal(t, "append", fd.Name)
	require.Len(t, fd.Params, 2)
	require.Equal(t, "log", fd.Params[0].Name)
	require.Equal(t, "data", fd.Params[1].Name)
}

func TestParsePredicateSimpleBody(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate always() { true }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 1)

	pd := spec.Declarations[0].(*ast.PredicateDecl)
	require.Equal(t, "always", pd.Name)
	require.Empty(t, pd.Params)

	lit := pd.Body.(*ast.LiteralExpr)
	require.Equal(t, "true", lit.Value)
	require.Equal(t, ast.LiteralBool, lit.Kind)
}

func TestParsePredicateWithParams(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate is_positive(x: int) { x > 0 }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[0].(*ast.PredicateDecl)
	require.Equal(t, "is_positive", pd.Name)
	require.Len(t, pd.Params, 1)
	require.Equal(t, "x", pd.Params[0].Name)
	require.Equal(t, "int", pd.Params[0].Type.Name)

	bin := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, ">", bin.Op)
	require.Equal(t, "x", bin.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "0", bin.Right.(*ast.LiteralExpr).Value)
}

func TestParsePredicateMultipleParams(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate in_bounds(x: int, lo: int, hi: int) { x >= lo and x <= hi }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[0].(*ast.PredicateDecl)
	require.Equal(t, "in_bounds", pd.Name)
	require.Len(t, pd.Params, 3)

	// x >= lo and x <= hi → BinaryExpr(and, BinaryExpr(>=, x, lo), BinaryExpr(<=, x, hi))
	and := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "and", and.Op)

	left := and.Left.(*ast.BinaryExpr)
	require.Equal(t, ">=", left.Op)
	require.Equal(t, "x", left.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "lo", left.Right.(*ast.IdentExpr).Name)

	right := and.Right.(*ast.BinaryExpr)
	require.Equal(t, "<=", right.Op)
	require.Equal(t, "x", right.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "hi", right.Right.(*ast.IdentExpr).Name)
}

func TestParseExprMulBindsTighterThanAdd(t *testing.T) {
	// arrange — 1 + 2 * 3 should parse as 1 + (2 * 3)
	source := `spec "test" {
                predicate p() { 1 + 2 * 3 }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	add := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "+", add.Op)
	require.Equal(t, "1", add.Left.(*ast.LiteralExpr).Value)

	mul := add.Right.(*ast.BinaryExpr)
	require.Equal(t, "*", mul.Op)
	require.Equal(t, "2", mul.Left.(*ast.LiteralExpr).Value)
	require.Equal(t, "3", mul.Right.(*ast.LiteralExpr).Value)
}

func TestParseExprComparison(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate p(a: int, b: int) { a == b }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	eq := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "==", eq.Op)
	require.Equal(t, "a", eq.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "b", eq.Right.(*ast.IdentExpr).Name)
}

func TestParseExprUnaryMinus(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate p(x: int) { -x }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	u := pd.Body.(*ast.UnaryExpr)
	require.Equal(t, "-", u.Op)
	require.Equal(t, "x", u.Operand.(*ast.IdentExpr).Name)
}

func TestParseExprUnaryNot(t *testing.T) {
	// arrange
	source := `spec "test" {
                predicate p(x: bool) { not x }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	u := pd.Body.(*ast.UnaryExpr)
	require.Equal(t, "not", u.Op)
	require.Equal(t, "x", u.Operand.(*ast.IdentExpr).Name)
}

func TestParseExprNestedParens(t *testing.T) {
	// arrange — (1 + 2) * 3
	source := `spec "test" {
                predicate p() { (1 + 2) * 3 }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	mul := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "*", mul.Op)
	require.Equal(t, "3", mul.Right.(*ast.LiteralExpr).Value)

	add := mul.Left.(*ast.BinaryExpr)
	require.Equal(t, "+", add.Op)
	require.Equal(t, "1", add.Left.(*ast.LiteralExpr).Value)
	require.Equal(t, "2", add.Right.(*ast.LiteralExpr).Value)
}

func TestParseExprAndBindsTighterThanOr(t *testing.T) {
	// arrange — a and b or c should parse as (a and b) or c
	source := `spec "test" {
                predicate p(a: bool, b: bool, c: bool) { a and b or c }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	or := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "or", or.Op)
	require.Equal(t, "c", or.Right.(*ast.IdentExpr).Name)

	and := or.Left.(*ast.BinaryExpr)
	require.Equal(t, "and", and.Op)
	require.Equal(t, "a", and.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "b", and.Right.(*ast.IdentExpr).Name)
}

func TestParseExprLeftAssociativity(t *testing.T) {
	// arrange — 1 - 2 - 3 should parse as (1 - 2) - 3
	source := `spec "test" {
                predicate p() { 1 - 2 - 3 }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	pd := spec.Declarations[0].(*ast.PredicateDecl)

	outer := pd.Body.(*ast.BinaryExpr)
	require.Equal(t, "-", outer.Op)
	require.Equal(t, "3", outer.Right.(*ast.LiteralExpr).Value)

	inner := outer.Left.(*ast.BinaryExpr)
	require.Equal(t, "-", inner.Op)
	require.Equal(t, "1", inner.Left.(*ast.LiteralExpr).Value)
	require.Equal(t, "2", inner.Right.(*ast.LiteralExpr).Value)
}

func TestParseExprAllLiteralTypes(t *testing.T) {
	// arrange — exercises int, float, string, bool literals
	source := `spec "test" {
                predicate p() { 42 }
                predicate q() { 3.14 }
                predicate r() { "hello" }
                predicate s() { false }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 4)

	p := spec.Declarations[0].(*ast.PredicateDecl).Body.(*ast.LiteralExpr)
	require.Equal(t, "42", p.Value)
	require.Equal(t, ast.LiteralInt, p.Kind)

	q := spec.Declarations[1].(*ast.PredicateDecl).Body.(*ast.LiteralExpr)
	require.Equal(t, "3.14", q.Value)
	require.Equal(t, ast.LiteralFloat, q.Kind)

	r := spec.Declarations[2].(*ast.PredicateDecl).Body.(*ast.LiteralExpr)
	require.Equal(t, "hello", r.Value)
	require.Equal(t, ast.LiteralString, r.Kind)

	s := spec.Declarations[3].(*ast.PredicateDecl).Body.(*ast.LiteralExpr)
	require.Equal(t, "false", s.Value)
	require.Equal(t, ast.LiteralBool, s.Kind)
}

func TestParsePredicateErrorMissingName(t *testing.T) {
	// arrange
	source := `spec "test" { predicate () { true } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePredicateErrorMissingBrace(t *testing.T) {
	// arrange
	source := `spec "test" { predicate p() true }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePredicateTrailingCommaInParams(t *testing.T) {
	// arrange
	source := `spec "test" {
                  predicate in_range(x: int, lo: int, hi: int,) { x >= lo and x <= hi }
          }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[0].(*ast.PredicateDecl)
	require.Equal(t, "in_range", pd.Name)
	require.Len(t, pd.Params, 3)
	require.Equal(t, "x", pd.Params[0].Name)
	require.Equal(t, "lo", pd.Params[1].Name)
	require.Equal(t, "hi", pd.Params[2].Name)
}

func TestParsePredicateLeadingCommaInParams(t *testing.T) {
	// arrange — leading comma is malformed; parser must not hang
	source := `spec "test" { predicate p(, x: int) { true } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertyContractualBareType(t *testing.T) {
	// arrange — commutative property from 'math' spec
	source := `spec "math" {
                func add(a: int, b: int) -> int

                property commutative forall(a: int, b: int) {
                        add(a, b) == add(b, a)
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)
	require.Len(t, spec.Declarations, 2)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Equal(t, "commutative", pd.Name)
	require.Equal(t, ast.Contractual, pd.Shape)

	require.Len(t, pd.Forall.Vars, 2)
	require.Equal(t, "a", pd.Forall.Vars[0].Name)
	require.Equal(t, "int", pd.Forall.Vars[0].Type.Name)
	require.Nil(t, pd.Forall.Vars[0].Generator)
	require.Equal(t, "b", pd.Forall.Vars[1].Name)
	require.Equal(t, "int", pd.Forall.Vars[1].Type.Name)
	require.Nil(t, pd.Forall.Vars[1].Generator)

	require.Len(t, pd.Body, 1)
	assert := pd.Body[0].(*ast.AssertExpr)
	eq := assert.Expr.(*ast.BinaryExpr)
	require.Equal(t, "==", eq.Op)

	lhs := eq.Left.(*ast.CallExpr)
	require.Equal(t, "add", lhs.Func)
	require.Len(t, lhs.Args, 2)
	require.Equal(t, "a", lhs.Args[0].(*ast.IdentExpr).Name)
	require.Equal(t, "b", lhs.Args[1].(*ast.IdentExpr).Name)

	rhs := eq.Right.(*ast.CallExpr)
	require.Equal(t, "add", rhs.Func)
	require.Equal(t, "b", rhs.Args[0].(*ast.IdentExpr).Name)
	require.Equal(t, "a", rhs.Args[1].(*ast.IdentExpr).Name)
}

func TestParsePropertyRangeGen(t *testing.T) {
	// arrange
	source := `spec "test" {
                func square(n: int) -> int

                property square_non_negative forall(n: int in 1..100) {
                        square(n) >= 0
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Equal(t, "square_non_negative", pd.Name)
	require.Len(t, pd.Forall.Vars, 1)

	v := pd.Forall.Vars[0]
	require.Equal(t, "n", v.Name)
	require.Equal(t, "int", v.Type.Name)

	rg := v.Generator.(*ast.RangeGen)
	require.Equal(t, "1", rg.Lo.(*ast.LiteralExpr).Value)
	require.Equal(t, "100", rg.Hi.(*ast.LiteralExpr).Value)
}

func TestParsePropertyBuiltinGen(t *testing.T) {
	// arrange
	source := `spec "test" {
                func length(s: string) -> int

                property length_bounded forall(s: string in strings(1, 50)) {
                        length(s) >= 1
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Len(t, pd.Forall.Vars, 1)

	v := pd.Forall.Vars[0]
	require.Equal(t, "s", v.Name)
	require.Equal(t, "string", v.Type.Name)

	bg := v.Generator.(*ast.BuiltinGen)
	require.Equal(t, "strings", bg.Name)
	require.Len(t, bg.Args, 2)
	require.Equal(t, "1", bg.Args[0].(*ast.LiteralExpr).Value)
	require.Equal(t, "50", bg.Args[1].(*ast.LiteralExpr).Value)
}

func TestParsePropertyOneOfGen(t *testing.T) {
	// arrange
	source := `spec "test" {
                func classify(x: int) -> string

                property classifies_known forall(x: int in one_of(1, 2, 3)) {
                        classify(x) != "unknown"
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Len(t, pd.Forall.Vars, 1)

	v := pd.Forall.Vars[0]
	require.Equal(t, "x", v.Name)

	og := v.Generator.(*ast.OneOfGen)
	require.Len(t, og.Values, 3)
	require.Equal(t, "1", og.Values[0].(*ast.LiteralExpr).Value)
	require.Equal(t, "2", og.Values[1].(*ast.LiteralExpr).Value)
	require.Equal(t, "3", og.Values[2].(*ast.LiteralExpr).Value)
}

func TestParsePropertyMultipleForallVars(t *testing.T) {
	// arrange — mixed generator forms
	source := `spec "test" {
                func concat(a: string, b: string) -> string

                property concat_length forall(a: string in strings(0, 10), b: string in strings(0, 10)) {
                        length(concat(a, b)) == length(a) + length(b)
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Len(t, pd.Forall.Vars, 2)

	require.Equal(t, "a", pd.Forall.Vars[0].Name)
	bg0 := pd.Forall.Vars[0].Generator.(*ast.BuiltinGen)
	require.Equal(t, "strings", bg0.Name)

	require.Equal(t, "b", pd.Forall.Vars[1].Name)
	bg1 := pd.Forall.Vars[1].Generator.(*ast.BuiltinGen)
	require.Equal(t, "strings", bg1.Name)
}

func TestParsePropertySequential(t *testing.T) {
	// arrange — sequential property with let and require
	source := `spec "log" {
                type Log
                func append(log: Log, data: bytes) -> (uint, error)
                func read(log: Log, offset: uint) -> (bytes, error)

                property write_read forall(log: Log, data: bytes) {
                        let (offset, err) = append(log, data)
                        require err is ok
                        let (result, err2) = read(log, offset)
                        require err2 is ok
                        result == data
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[3].(*ast.PropertyDecl)
	require.Equal(t, "write_read", pd.Name)
	require.Equal(t, ast.Sequential, pd.Shape)

	require.Len(t, pd.Forall.Vars, 2)
	require.Equal(t, "log", pd.Forall.Vars[0].Name)
	require.Equal(t, "Log", pd.Forall.Vars[0].Type.Name)
	require.Equal(t, "data", pd.Forall.Vars[1].Name)
	require.Equal(t, "bytes", pd.Forall.Vars[1].Type.Name)

	require.Len(t, pd.Body, 5)

	let0 := pd.Body[0].(*ast.LetBinding)
	require.Equal(t, []string{"offset", "err"}, let0.Names)
	call0 := let0.Expr.(*ast.CallExpr)
	require.Equal(t, "append", call0.Func)

	req0 := pd.Body[1].(*ast.RequireStmt)
	isExpr0 := req0.Expr.(*ast.IsExpr)
	require.Equal(t, "err", isExpr0.Expr.(*ast.IdentExpr).Name)
	require.Equal(t, ast.IsOk, isExpr0.Target)

	let1 := pd.Body[2].(*ast.LetBinding)
	require.Equal(t, []string{"result", "err2"}, let1.Names)

	req1 := pd.Body[3].(*ast.RequireStmt)
	isExpr1 := req1.Expr.(*ast.IsExpr)
	require.Equal(t, ast.IsOk, isExpr1.Target)

	assert := pd.Body[4].(*ast.AssertExpr)
	eq := assert.Expr.(*ast.BinaryExpr)
	require.Equal(t, "==", eq.Op)
	require.Equal(t, "result", eq.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "data", eq.Right.(*ast.IdentExpr).Name)
}

func TestParsePropertyErrorMissingName(t *testing.T) {
	// arrange
	source := `spec "test" { property forall(x: int) { x > 0 } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertyErrorMissingForall(t *testing.T) {
	// arrange
	source := `spec "test" { property p (x: int) { x > 0 } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertyErrorMissingBrace(t *testing.T) {
	// arrange
	source := `spec "test" { property p forall(x: int) x > 0 }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertyErrorMalformedForallVar(t *testing.T) {
	// arrange — leading comma in forall vars, parser must not hang
	source := `spec "test" { property p forall(, x: int) { x > 0 } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertyWhereClause(t *testing.T) {
	// arrange
	source := `spec "test" {
                func div(a: int, b: int) -> int

                property div_identity forall(a: int, b: int) where b != 0 {
                        div(a * b, b) == a
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Equal(t, "div_identity", pd.Name)
	require.Equal(t, ast.Contractual, pd.Shape)

	require.NotNil(t, pd.Where)
	require.Equal(t, 4, pd.Where.Pos.Line)
	cond := pd.Where.Condition.(*ast.BinaryExpr)
	require.Equal(t, "!=", cond.Op)
	require.Equal(t, "b", cond.Left.(*ast.IdentExpr).Name)
	require.Equal(t, "0", cond.Right.(*ast.LiteralExpr).Value)

	require.Len(t, pd.Body, 1)
	assert := pd.Body[0].(*ast.AssertExpr)
	eq := assert.Expr.(*ast.BinaryExpr)
	require.Equal(t, "==", eq.Op)
}

func TestParsePropertyWhereWithSequentialBody(t *testing.T) {
	// arrange — where clause combined with let-bindings
	source := `spec "test" {
                  func div(a: int, b: int) -> (int, error)

                  property div_round_trip forall(a: int, b: int) where b != 0 {
                          let (result, err) = div(a * b, b)
                          require err is ok
                          result == a
                  }
          }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Equal(t, "div_round_trip", pd.Name)

	require.NotNil(t, pd.Where)
	cond := pd.Where.Condition.(*ast.BinaryExpr)
	require.Equal(t, "!=", cond.Op)

	require.Equal(t, ast.Sequential, pd.Shape)

	require.Len(t, pd.Body, 3)
	_ = pd.Body[0].(*ast.LetBinding)
	_ = pd.Body[1].(*ast.RequireStmt)
	_ = pd.Body[2].(*ast.AssertExpr)
}

func TestParsePropertyErrorMalformedWhere(t *testing.T) {
	// arrange
	source := `spec "test" { property p forall(x: int) where { x > 0 } }`

	// act
	_, errs := Parse(source, "test.assay")

	// assert
	require.NotEmpty(t, errs)
}

func TestParsePropertySingleLet(t *testing.T) {
	// arrange
	source := `spec "test" {
                func double(x: int) -> int

                property double_is_sum forall(x: int) {
                        let y = double(x)
                        y == x + x
                }
        }`

	// act
	spec, errs := Parse(source, "test.assay")

	// assert
	require.Empty(t, errs)

	pd := spec.Declarations[1].(*ast.PropertyDecl)
	require.Equal(t, "double_is_sum", pd.Name)
	require.Equal(t, ast.Sequential, pd.Shape)

	require.Len(t, pd.Body, 2)

	let0 := pd.Body[0].(*ast.LetBinding)
	require.Equal(t, []string{"y"}, let0.Names)
	call := let0.Expr.(*ast.CallExpr)
	require.Equal(t, "double", call.Func)
	require.Len(t, call.Args, 1)
	require.Equal(t, "x", call.Args[0].(*ast.IdentExpr).Name)

	assert := pd.Body[1].(*ast.AssertExpr)
	eq := assert.Expr.(*ast.BinaryExpr)
	require.Equal(t, "==", eq.Op)
	require.Equal(t, "y", eq.Left.(*ast.IdentExpr).Name)
}
