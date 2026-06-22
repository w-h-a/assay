package resolver

import (
	"fmt"

	bindingast "github.com/w-h-a/assay/internal/binding/ast"
)

// Error is a resolution failure with the source position of the
// binding mapping that could not be resolved against the target package.
type Error struct {
	Message string
	Pos     bindingast.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}
