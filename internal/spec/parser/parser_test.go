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
