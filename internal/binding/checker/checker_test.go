package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
	bindingast "github.com/w-h-a/assay/internal/binding/ast"
	bindingparser "github.com/w-h-a/assay/internal/binding/parser"
	specast "github.com/w-h-a/assay/internal/spec/ast"
	specchecker "github.com/w-h-a/assay/internal/spec/checker"
	specparser "github.com/w-h-a/assay/internal/spec/parser"
)

func TestCheckValidLogBinding(t *testing.T) {
	// arrange — pitch's log spec + matching binding
	spec := validatedSpec(t, `spec "log" {
		type Log
		func new_log() -> Log
		func append(log: Log, value: bytes) -> (uint, error)
		func read(log: Log, offset: uint) -> (bytes, error)
	}`)
	binding := parsedBinding(t, `bind "log" {
		target go
		package "github.com/w-h-a/tally/internal/log"
		type Log     = log.Log
		func new_log = log.New
		func append  = Log.Append
		func read    = Log.ReadAt
	}`)

	// act
	validated, errs := Check(binding, spec)

	// assert
	require.Empty(t, errs)
	require.Equal(t, binding, validated.Binding)
}

func TestCheckEmptyBindingAgainstEmptySpec(t *testing.T) {
	// arrange
	spec := validatedSpec(t, `spec "empty" {}`)
	binding := parsedBinding(t, `bind "empty" {}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckBindingNameMismatch(t *testing.T) {
	// arrange — binding name does not match spec name
	spec := validatedSpec(t, `spec "log" {}`)
	binding := parsedBinding(t, `bind "math" {}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "binding name")
	require.Contains(t, errs[0].Message, `"math"`)
	require.Contains(t, errs[0].Message, `"log"`)
	require.Equal(t, 1, errs[0].Pos.Line)
}

func TestCheckTypeMappingReferencesUndeclaredType(t *testing.T) {
	// arrange — spec has no type Widget
	spec := validatedSpec(t, `spec "x" { type Log }`)
	binding := parsedBinding(t, `bind "x" {
		type Widget = pkg.Widget
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undeclared spec type")
	require.Contains(t, errs[0].Message, `"Widget"`)
	require.Equal(t, 2, errs[0].Pos.Line)
}

func TestCheckTypeMappingReferencesDeclaredType(t *testing.T) {
	// arrange
	spec := validatedSpec(t, `spec "x" { type Log }`)
	binding := parsedBinding(t, `bind "x" {
		type Log = pkg.Log
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckFuncMappingReferencesUndeclaredFunc(t *testing.T) {
	// arrange — spec has no func read
	spec := validatedSpec(t, `spec "x" {
		type Log
		func new_log() -> Log
	}`)
	binding := parsedBinding(t, `bind "x" {
		func read = pkg.Read
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undeclared spec func")
	require.Contains(t, errs[0].Message, `"read"`)
	require.Equal(t, 2, errs[0].Pos.Line)
}

func TestCheckFuncMappingReferencesDeclaredFunc(t *testing.T) {
	// arrange
	spec := validatedSpec(t, `spec "x" {
		type Log
		func new_log() -> Log
	}`)
	binding := parsedBinding(t, `bind "x" {
		func new_log = pkg.New
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Empty(t, errs)
}

func TestCheckTypeMappingDoesNotResolveAgainstFunc(t *testing.T) {
	// arrange — spec has a func 'thing' but no type 'thing'
	spec := validatedSpec(t, `spec "x" { func thing() -> int }`)
	binding := parsedBinding(t, `bind "x" {
		type thing = pkg.Thing
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert -- a type mapping must not resolve against a func name
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undeclared spec type")
	require.Contains(t, errs[0].Message, `"thing"`)
}

func TestCheckFuncMappingDoesNotResolveAgainstType(t *testing.T) {
	// arrange — spec has a type 'Thing' but no func 'Thing'
	spec := validatedSpec(t, `spec "x" { type Thing }`)
	binding := parsedBinding(t, `bind "x" {
		func Thing = pkg.Thing
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert -- a func mapping must not resolve against a type name
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "undeclared spec func")
	require.Contains(t, errs[0].Message, `"Thing"`)
}

func TestCheckCollectsMultipleErrors(t *testing.T) {
	// arrange — name mismatch + undeclared type + undeclared func
	spec := validatedSpec(t, `spec "log" { type Log }`)
	binding := parsedBinding(t, `bind "math" {
		type Widget = pkg.Widget
		func compute = pkg.Compute
	}`)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Len(t, errs, 3)
}

func TestCheckIgnoresEmptyNamesFromParserRecovery(t *testing.T) {
	// arrange — empty binding name and empty mapping SpecNames, as the
	// parser produces on recovery from missing identifiers
	spec := &specast.ValidatedSpec{Spec: &specast.SpecDecl{Name: "x"}}
	binding := &bindingast.BindingDecl{
		Name: "",
		Pos:  bindingast.Position{Line: 1, Column: 1},
		TypeMappings: []bindingast.TypeMapping{
			{SpecName: "", Pos: bindingast.Position{Line: 2, Column: 3}},
		},
		FuncMappings: []bindingast.FuncMapping{
			{SpecName: "", Pos: bindingast.Position{Line: 3, Column: 3}},
		},
	}

	// act
	_, errs := Check(binding, spec)

	// assert -- the binding parser already reported the missing identifiers
	require.Empty(t, errs)
}

func TestCheckErrorPositionIsBindingPosition(t *testing.T) {
	// arrange — confirm the error points at the binding declaration, not the spec
	spec := validatedSpec(t, `spec "x" {}`)
	source := "bind \"x\" {\n\ttype Log = pkg.Log\n}"
	binding := parsedBinding(t, source)

	// act
	_, errs := Check(binding, spec)

	// assert
	require.Len(t, errs, 1)
	require.Equal(t, 2, errs[0].Pos.Line)
}

// validatedSpec parses a spec source, type-checks it, and fails the test if
// either step produces errors.
func validatedSpec(t *testing.T, source string) *specast.ValidatedSpec {
	t.Helper()

	spec, parseErrs := specparser.Parse(source, "test.assay")
	require.Empty(t, parseErrs, "spec parse errors: %v", parseErrs)

	validated, checkErrs := specchecker.Check(spec)
	require.Empty(t, checkErrs, "spec check errors: %v", checkErrs)

	return validated
}

// parsedBinding parses a binding source and fails the test if parsing produces errors.
func parsedBinding(t *testing.T, source string) *bindingast.BindingDecl {
	t.Helper()

	binding, errs := bindingparser.Parse(source, "test.bind")
	require.Empty(t, errs, "binding parse errors: %v", errs)

	return binding
}
