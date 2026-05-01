package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/w-h-a/assay/internal/spec/ast"
	"github.com/w-h-a/assay/internal/spec/parser"
)

func TestCheckEmptySpec(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "empty" {}`)

	// act
	validated, errs := Check(spec)

	// assert
	require.Empty(t, errs)
	require.Equal(t, spec, validated.Spec)
}

func TestCheckRegistersAllDeclarationKinds(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                type Log
                func new_log() -> Log
                predicate non_empty(v: bytes) { len(v) > 0 }
                property identity forall(a: int) { a == a }
        }`)

	// act
	validated, errs := Check(spec)

	// assert
	require.Empty(t, errs)
	require.Equal(t, spec, validated.Spec)
}

func TestCheckDuplicateTypeNames(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                type Log
                type Log
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Log")
	require.Contains(t, errs[0].Message, "already declared")
	require.Equal(t, 3, errs[0].Pos.Line)
}

func TestCheckDuplicateFuncNames(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
				type Log
                func new_log() -> Log
                func new_log() -> Log
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "new_log")
	require.Contains(t, errs[0].Message, "already declared")
}

func TestCheckDuplicatePredicateNames(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                predicate p(x: int) { x > 0 }
                predicate p(y: int) { y > 0 }
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "p")
	require.Contains(t, errs[0].Message, "already declared")
}

func TestCheckDuplicatePropertyNames(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                property p forall(a: int) { a == a }
                property p forall(b: int) { b == b }
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "p")
	require.Contains(t, errs[0].Message, "already declared")
}

func TestCheckDuplicateAcrossKinds(t *testing.T) {
	// arrange -- type and func sharing the same name
	spec := parseValid(t, `spec "test" {
                type Log
                func Log() -> int
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Log")
	require.Contains(t, errs[0].Message, "already declared")
}

func TestCheckMultipleDuplicates(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                type Log
                type Log
                func read() -> int
                func read() -> int
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 2)
	require.Contains(t, errs[0].Message, "Log")
	require.Contains(t, errs[1].Message, "read")
}

func TestCheckDuplicateErrorReferencesOriginalPosition(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                type Log
                type Log
        }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "2:") // original at line 2
	require.Equal(t, 3, errs[0].Pos.Line)      // error points to line 3
}

func TestCheckFullLogSpec(t *testing.T) {
	// arrange -- pitch log spec with no duplicates
	spec := parseValid(t, `spec "log" {
                type Log
                func new_log() -> Log
                func append(log: Log, value: bytes) -> (uint, error)
                func read(log: Log, offset: uint) -> (bytes, error)

                predicate non_empty(v: bytes) {
                        len(v) > 0
                }

                property read_after_write
                        forall(value: bytes)
                        where non_empty(value)
                {
                        let log = new_log()
                        let (offset, err) = append(log, value)
                        require err is ok
                        let (result, err2) = read(log, offset)
                        require err2 is ok
                        result == value
                }

                property append_monotonic
                        forall(a: bytes, b: bytes)
                        where non_empty(a) and non_empty(b)
                {
                        let log = new_log()
                        let (off_a, _) = append(log, a)
                        let (off_b, _) = append(log, b)
                        off_b > off_a
                }
        }`)

	// act
	validated, errs := Check(spec)

	// assert
	require.Empty(t, errs)
	require.Equal(t, spec, validated.Spec)
}

func TestCheckIgnoresEmptyNamesFromParserRecovery(t *testing.T) {
	// arrange -- two declarations with empty names, as the parser
	// produces on recovery from missing identifiers (e.g., "type {")
	spec := &ast.SpecDecl{
		Name: "test",
		Declarations: []ast.Decl{
			&ast.TypeDecl{Name: "", Pos: ast.Position{Line: 2, Column: 17}},
			&ast.TypeDecl{Name: "", Pos: ast.Position{Line: 3, Column: 17}},
		},
	}

	// act
	_, errs := Check(spec)

	// assert -- checker should not report duplicates for empty names;
	// the parser already reported the missing identifiers.
	require.Empty(t, errs)
}

func TestCheckValidStructFieldTypes(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          offset: uint,
                          data: bytes
                  }
          }`)

	// act
	validated, errs := Check(spec)

	// assert
	require.Empty(t, errs)
	require.Equal(t, spec, validated.Spec)
}

func TestCheckUndefinedStructFieldType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          data: Blob
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Blob")
	require.Contains(t, errs[0].Message, "undefined type")
	require.Equal(t, 3, errs[0].Pos.Line)
}

func TestCheckStructFieldResolvesToDeclaredType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Log
                  type Entry {
                          source: Log,
                          offset: uint
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckUndefinedFuncParamType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  func read(log: Store) -> bytes
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Store")
	require.Contains(t, errs[0].Message, "undefined type")
	require.Equal(t, 2, errs[0].Pos.Line)
}

func TestCheckUndefinedFuncReturnType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  func create() -> Widget
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Widget")
	require.Contains(t, errs[0].Message, "undefined type")
}

func TestCheckValidFuncSignatureWithBuiltins(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  func compute(a: int, b: float) -> (uint, error)
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckValidFuncSignatureWithDeclaredType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Log
                  func new_log() -> Log
                  func append(log: Log, value: bytes) -> (uint, error)
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckParameterizedTypeInnerResolution(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry
                  func entries() -> list[Entry]
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckUndefinedInnerTypeInParameterized(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  func entries() -> list[Widget]
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Widget")
	require.Contains(t, errs[0].Message, "undefined type")
}

func TestCheckNestedParameterizedTypes(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          items: option[list[int]]
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckUndefinedInNestedParameterizedType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          items: option[list[Widget]]
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Widget")
	require.Contains(t, errs[0].Message, "undefined type")
}

func TestCheckMapParameterizedType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  func get(store: map[string, bytes], key: string) -> option[bytes]
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckMultipleUndefinedTypes(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          data: Blob
                  }
                  func read(log: Store) -> Widget
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 3)
}

func TestCheckBareParameterizedType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          items: list
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "list")
	require.Contains(t, errs[0].Message, "expects 1 type parameter(s)")
}

func TestCheckScalarWithTypeParameters(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          data: bool[string]
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "bool")
	require.Contains(t, errs[0].Message, "does not accept type parameters")
}

func TestCheckWrongTypeParameterCount(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          data: map[string]
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "map")
	require.Contains(t, errs[0].Message, "expects 2 type parameter(s), got 1")
}

func TestCheckCorrectParameterizedTypeUsage(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Entry {
                          items: list[int],
                          tags: set[string],
                          metadata: map[string, bytes],
                          parent: option[error],
                          count: int,
                          flag: bool
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckUserDefinedTypeWithTypeParameters(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                    type Entry
                    type Container {
                            data: Entry[string]
                    }
            }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Entry")
	require.Contains(t, errs[0].Message, "does not accept type parameters")
}

// parseValid parses source and fails the test if parsing produces errors.
func parseValid(t *testing.T, source string) *ast.SpecDecl {
	t.Helper()

	spec, errs := parser.Parse(source, "test.assay")
	require.Empty(t, errs, "parse errors: %v", errs)

	return spec
}
