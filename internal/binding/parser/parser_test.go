package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLogBind(t *testing.T) {
	// arrange — exact bind from the pitch
	source := `bind "log" {
		target go
		package "github.com/w-h-a/tally/internal/log"
		type Log     = log.Log
		func new_log = log.New
		func append  = Log.Append
		func read    = Log.ReadAt
	}`

	// act
	decl, errs := Parse(source, "log.bind")

	// assert
	require.Empty(t, errs)
	require.Equal(t, "log", decl.Name)
	require.Equal(t, "go", decl.Target)
	require.Equal(t, "github.com/w-h-a/tally/internal/log", decl.PackagePath)
	require.Equal(t, 1, decl.Pos.Line)
	require.Equal(t, 1, decl.Pos.Column)

	require.Len(t, decl.TypeMappings, 1)
	require.Equal(t, "Log", decl.TypeMappings[0].SpecName)
	require.Equal(t, "log", decl.TypeMappings[0].Qualifier)
	require.Equal(t, "Log", decl.TypeMappings[0].Name)

	require.Len(t, decl.FuncMappings, 3)

	require.Equal(t, "new_log", decl.FuncMappings[0].SpecName)
	require.Equal(t, "log", decl.FuncMappings[0].Qualifier)
	require.Equal(t, "New", decl.FuncMappings[0].Name)

	require.Equal(t, "append", decl.FuncMappings[1].SpecName)
	require.Equal(t, "Log", decl.FuncMappings[1].Qualifier)
	require.Equal(t, "Append", decl.FuncMappings[1].Name)

	require.Equal(t, "read", decl.FuncMappings[2].SpecName)
	require.Equal(t, "Log", decl.FuncMappings[2].Qualifier)
	require.Equal(t, "ReadAt", decl.FuncMappings[2].Name)
}

func TestParseEmptyBind(t *testing.T) {
	// arrange
	source := `bind "empty" {}`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.Empty(t, errs)
	require.Equal(t, "empty", decl.Name)
	require.Empty(t, decl.Target)
	require.Empty(t, decl.PackagePath)
	require.Empty(t, decl.TypeMappings)
	require.Empty(t, decl.FuncMappings)
}

func TestParseMultipleTypeMappings(t *testing.T) {
	// arrange
	source := `bind "multi" {
		type Log   = log.Log
		type Entry = log.Entry
	}`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.Empty(t, errs)
	require.Len(t, decl.TypeMappings, 2)
	require.Equal(t, "Log", decl.TypeMappings[0].SpecName)
	require.Equal(t, "Entry", decl.TypeMappings[1].SpecName)
}

func TestParseFuncMappingPackageForm(t *testing.T) {
	// arrange — pkg.Func form
	source := `bind "x" {
		func new_log = log.New
	}`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.Empty(t, errs)
	require.Len(t, decl.FuncMappings, 1)
	require.Equal(t, "new_log", decl.FuncMappings[0].SpecName)
	require.Equal(t, "log", decl.FuncMappings[0].Qualifier)
	require.Equal(t, "New", decl.FuncMappings[0].Name)
}

func TestParseFuncMappingMethodForm(t *testing.T) {
	// arrange — Type.Method form parses to the same shape as pkg.Func;
	// the resolver classifies based on the spec's first parameter.
	source := `bind "x" {
		func append = Log.Append
	}`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.Empty(t, errs)
	require.Len(t, decl.FuncMappings, 1)
	require.Equal(t, "append", decl.FuncMappings[0].SpecName)
	require.Equal(t, "Log", decl.FuncMappings[0].Qualifier)
	require.Equal(t, "Append", decl.FuncMappings[0].Name)
}

func TestParsePositionTracking(t *testing.T) {
	// arrange
	source := "bind \"x\" {\n\ttype Log = log.Log\n}"

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.Empty(t, errs)
	require.Equal(t, "test.bind", decl.TypeMappings[0].Pos.File)
	require.Equal(t, 2, decl.TypeMappings[0].Pos.Line)
}

func TestParseErrorMissingBind(t *testing.T) {
	// arrange
	source := `"log" {}`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Message, "expected bind")
}

func TestParseErrorMissingName(t *testing.T) {
	// arrange
	source := `bind {}`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Message, "expected STRING_LIT")
}

func TestParseErrorMissingOpenBrace(t *testing.T) {
	// arrange
	source := `bind "x" target go`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Equal(t, "x", decl.Name)
}

func TestParseErrorMissingCloseBrace(t *testing.T) {
	// arrange
	source := `bind "x" {`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorMissingTargetValue(t *testing.T) {
	// arrange
	source := `bind "x" { target }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorMissingPackageValue(t *testing.T) {
	// arrange
	source := `bind "x" { package }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorTypeMissingAssign(t *testing.T) {
	// arrange
	source := `bind "x" { type Log log.Log }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorTypeMissingDottedRef(t *testing.T) {
	// arrange — qualifier present but no dot-name
	source := `bind "x" { type Log = log }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorFuncMissingName(t *testing.T) {
	// arrange
	source := `bind "x" { func = log.New }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
}

func TestParseErrorUnknownStatement(t *testing.T) {
	// arrange
	source := `bind "x" { foo bar }`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Message, "expected target, package, type, or func")
}

func TestParseErrorRecoveryContinuesAfterMalformed(t *testing.T) {
	// arrange — malformed type decl followed by valid func decl;
	// recovery should resync to FUNC and parse the rest.
	source := `bind "x" {
		type Log =
		func read = log.Read
	}`

	// act
	decl, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Len(t, decl.FuncMappings, 1)
	require.Equal(t, "read", decl.FuncMappings[0].SpecName)
}

func TestParseErrorTrailingTokens(t *testing.T) {
	// arrange — anything after the closing brace is invalid
	source := `bind "x" {} junk`

	// act
	_, errs := Parse(source, "test.bind")

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Message, "unexpected token")
}
