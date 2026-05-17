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

func TestCheckPredicateParamTypesResolve(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Log
                  predicate has_entries(log: Log, count: uint) { count > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPredicateUndefinedParamType(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  predicate valid(v: Widget) { v > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Widget")
	require.Contains(t, errs[0].Message, "undefined type")
}

func TestCheckPredicateRejectsSpecFuncCall(t *testing.T) {
	// arrange
	spec := parseValid(t, `spec "test" {
                  type Log
                  func new_log() -> Log
                  predicate bad(x: int) { new_log() > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "new_log")
	require.Contains(t, errs[0].Message, "cannot call function")
	require.Contains(t, errs[0].Message, "predicate body")
}

func TestCheckPredicateAllowsBuiltinCall(t *testing.T) {
	// arrange — len is not a spec-declared function, so it passes
	spec := parseValid(t, `spec "test" {
                  predicate non_empty(v: bytes) { len(v) > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPredicateRejectsNestedSpecFuncCall(t *testing.T) {
	// arrange — spec func call nested inside a binary expression
	spec := parseValid(t, `spec "test" {
                  func compute(x: int) -> int
                  predicate bad(x: int) { compute(x) > 0 and x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "compute")
	require.Contains(t, errs[0].Message, "cannot call function")
}

func TestCheckPredicateAllowsPredicateCall(t *testing.T) {
	// arrange — predicate calling another predicate is safe (both are pure)
	spec := parseValid(t, `spec "test" {
                    predicate positive(x: int) { x > 0 }
                    predicate valid(x: int) { positive(x) }
            }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPredicateBodyMustBeBool(t *testing.T) {
	// arrange — body is int, not bool
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { x + 1 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "predicate body must be a boolean expression")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckPredicateBodyBoolPasses(t *testing.T) {
	// arrange — body is bool
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPredicateParamAvailableInBody(t *testing.T) {
	// arrange — param used in body resolves
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int, y: int) { x + y > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckUndefinedIdentifierInExpression(t *testing.T) {
	// arrange — y is not a param
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { y > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undefined identifier")
	require.Contains(t, errs[0].Message, `"y"`)
}

func TestCheckArithmeticRequiresNumeric(t *testing.T) {
	// arrange — string in arithmetic
	spec := parseValid(t, `spec "test" {
                  predicate p(x: string) { x + x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires numeric operands")
	require.Contains(t, errs[0].Message, `"string"`)
}

func TestCheckComparisonMismatchedTypes(t *testing.T) {
	// arrange — int == string
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { x == "hello" }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires matching types")
	require.Contains(t, errs[0].Message, `"int"`)
	require.Contains(t, errs[0].Message, `"string"`)
}

func TestCheckComparisonCrossNumericAllowed(t *testing.T) {
	// arrange — uint compared with int literal
	spec := parseValid(t, `spec "test" {
                  predicate p(x: uint) { x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckLogicalRequiresBool(t *testing.T) {
	// arrange — int used with and
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { x and true }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires bool operands")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckUnaryNotRequiresBool(t *testing.T) {
	// arrange — not applied to int
	spec := parseValid(t, `spec "test" {
                  predicate p(x: int) { not x }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires bool operand")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckUnaryMinusRequiresNumeric(t *testing.T) {
	// arrange — negate a bool
	spec := parseValid(t, `spec "test" {
                  predicate p(x: bool) { -x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires numeric operand")
	require.Contains(t, errs[0].Message, `"bool"`)
}

func TestCheckPropertyForallVarInScope(t *testing.T) {
	// arrange — forall var used in assertion
	spec := parseValid(t, `spec "test" {
                  property p forall(x: int) { x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPropertyAssertionMustBeBool(t *testing.T) {
	// arrange — assertion is int, not bool
	spec := parseValid(t, `spec "test" {
                  property p forall(x: int) { x + 1 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "property assertion must be a boolean expression")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckPropertyLetBindingInScope(t *testing.T) {
	// arrange — let binding used in later assertion
	spec := parseValid(t, `spec "test" {
                  property p forall(x: int) {
                          let y = x + 1
                          y > 0
                  }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckPropertyUndefinedIdentifier(t *testing.T) {
	// arrange — z not in scope
	spec := parseValid(t, `spec "test" {
                  property p forall(x: int) { z > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undefined identifier")
	require.Contains(t, errs[0].Message, `"z"`)
}

func TestCheckPropertyWhereClauseMustBeBool(t *testing.T) {
	// arrange — where clause is int, not bool
	spec := parseValid(t, `spec "test" {
                  property p forall(x: int) where x + 1 { x > 0 }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "where clause must be a boolean expression")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckPropertyForallVarTypeResolution(t *testing.T) {
	// arrange — forall var references undefined type
	spec := parseValid(t, `spec "test" {
                  property p forall(x: Widget) { x == x }
          }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undefined type")
	require.Contains(t, errs[0].Message, `"Widget"`)
}

func TestCheckPropertyRequireMustBeBool(t *testing.T) {
	// arrange — require condition is int, not bool
	spec := parseValid(t, `spec "test" {
                    property p forall(x: int) {
                            require x + 1
                            x > 0
                    }
            }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "require condition must be a boolean expression")
	require.Contains(t, errs[0].Message, `"int"`)
}

func TestCheckArithmeticMismatchedNumericTypes(t *testing.T) {
	// arrange — int + float is a type mismatch
	spec := parseValid(t, `spec "test" {
                    predicate p(x: int, y: float) { x + y > 0 }
            }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "requires matching numeric types")
	require.Contains(t, errs[0].Message, `"int"`)
	require.Contains(t, errs[0].Message, `"float"`)
}

func TestCheckPropertyWhereRejectsSpecFuncCall(t *testing.T) {
	// arrange — where clause calls a spec-declared function
	spec := parseValid(t, `spec "test" {
                    type Log
                    func new_log() -> Log
                    predicate valid(x: int) { x > 0 }
                    property p forall(x: int) where new_log() is ok { x > 0 }
            }`)

	// act
	_, errs := Check(spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "cannot call function")
	require.Contains(t, errs[0].Message, `"new_log"`)
	require.Contains(t, errs[0].Message, "where clause")
}
