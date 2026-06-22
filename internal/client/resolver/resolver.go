package resolver

import (
	"context"

	bindingast "github.com/w-h-a/assay/internal/binding/ast"
	specast "github.com/w-h-a/assay/internal/spec/ast"
)

// Resolver is the port interface for resolving a validated binding's
// cross-references against the target language's type system.
type Resolver interface {
	Resolve(ctx context.Context, binding *bindingast.ValidatedBinding, spec *specast.ValidatedSpec) (*ResolvedBinding, []Error)
}
