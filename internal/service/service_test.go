package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckValidSpec(t *testing.T) {
	// arrange
	path := writeSpec(t, `spec "math" {
                func add(a: int, b: int) -> int

                property commutative forall(a: int, b: int) {
                        add(a, b) == add(b, a)
                }
        }`)
	svc := New()

	// act
	errs := svc.Check(path)

	// assert
	require.Empty(t, errs)
}

func TestCheckTypeError(t *testing.T) {
	// arrange
	path := writeSpec(t, `spec "test" {
                predicate p(x: int) { x + "hello" > 0 }
        }`)
	svc := New()

	// act
	errs := svc.Check(path)

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Error(), "requires numeric operands")
	require.Contains(t, errs[0].Error(), ".assay:")
}

func TestCheckParseError(t *testing.T) {
	// arrange
	path := writeSpec(t, `spec "test" {
                type
        }`)
	svc := New()

	// act
	errs := svc.Check(path)

	// assert
	require.NotEmpty(t, errs)
	require.Contains(t, errs[0].Error(), ".assay:")
}

func TestCheckFileNotFound(t *testing.T) {
	// arrange
	svc := New()

	// act
	errs := svc.Check("/nonexistent/spec.assay")

	// assert
	require.Len(t, errs, 1)
	require.ErrorIs(t, errs[0], os.ErrNotExist)
}

func writeSpec(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.assay")
	err := os.WriteFile(path, []byte(source), 0o644)
	require.NoError(t, err)

	return path
}
