package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/w-h-a/assay/internal/handler/cli"
	"github.com/w-h-a/assay/internal/service"
)

func TestVersion(t *testing.T) {
	// arrange
	h := cli.New(service.New())
	out := &bytes.Buffer{}

	// act
	err := h.Version(out)

	// assert
	require.NoError(t, err)
	require.Contains(t, out.String(), "assay")
}

func TestCheckValidSpec(t *testing.T) {
	// arrange
	path := writeSpec(t, `spec "math" {
                func add(a: int, b: int) -> int

                property commutative forall(a: int, b: int) {
                        add(a, b) == add(b, a)
                }
        }`)
	h := cli.New(service.New())
	stderr := &bytes.Buffer{}

	// act
	err := h.Check(stderr, path)

	// assert
	require.NoError(t, err)
	require.Empty(t, stderr.String())
}

func TestCheckInvalidSpec(t *testing.T) {
	// arrange
	path := writeSpec(t, `spec "test" {
                predicate p(x: int) { x + "hello" > 0 }
        }`)
	h := cli.New(service.New())
	stderr := &bytes.Buffer{}

	// act
	err := h.Check(stderr, path)

	// assert
	require.Error(t, err)
	require.Contains(t, stderr.String(), "requires numeric operands")
}

func TestCheckFileNotFound(t *testing.T) {
	// arrange
	h := cli.New(service.New())
	stderr := &bytes.Buffer{}

	// act
	err := h.Check(stderr, "/nonexistent/spec.assay")

	// assert
	require.Error(t, err)
	require.Contains(t, stderr.String(), "no such file")
}

func writeSpec(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.assay")
	err := os.WriteFile(path, []byte(source), 0o644)
	require.NoError(t, err)

	return path
}
