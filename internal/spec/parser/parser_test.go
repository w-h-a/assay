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
