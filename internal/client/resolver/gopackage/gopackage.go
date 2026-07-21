package gopackage

import (
	"context"
	"fmt"
	"go/types"

	bindingast "github.com/w-h-a/assay/internal/binding/ast"
	"github.com/w-h-a/assay/internal/client/resolver"
	specast "github.com/w-h-a/assay/internal/spec/ast"
)

// goResolver resolves a validated binding against a target Go package
// using go/packages.
type goResolver struct {
	options resolver.Options
}

func New(opts ...resolver.Option) (resolver.Resolver, error) {
	options := resolver.NewOptions(opts...)

	r := &goResolver{
		options: options,
	}

	return r, nil
}

// Resolve loads the binding's target package and resolves its type
// mappings against it.
func (r *goResolver) Resolve(
	ctx context.Context,
	binding *bindingast.ValidatedBinding,
	spec *specast.ValidatedSpec,
) (*resolver.ResolvedBinding, []resolver.Error) {
	decl := binding.Binding

	resolved := &resolver.ResolvedBinding{
		Funcs: map[string]resolver.ResolvedFunc{},
		Types: map[string]resolver.ResolvedType{},
	}

	pkg, err := loadPackage(ctx, decl.PackagePath)
	if err != nil {
		return resolved, []resolver.Error{{Message: err.Error(), Pos: decl.Pos}}
	}

	var errs []resolver.Error

	for _, m := range decl.TypeMappings {
		obj := pkg.Types.Scope().Lookup(m.Name)
		if obj == nil {
			errs = append(errs, resolver.Error{
				Message: fmt.Sprintf("type %q not found in package %q", m.Name, decl.PackagePath),
				Pos:     m.Pos,
			})
			continue
		}

		if _, ok := obj.(*types.TypeName); !ok {
			errs = append(errs, resolver.Error{
				Message: fmt.Sprintf("%q in package %q is not a type", m.Name, decl.PackagePath),
				Pos:     m.Pos,
			})
			continue
		}

		resolved.Types[m.SpecName] = resolver.ResolvedType{
			PackagePath: pkg.PkgPath,
			Name:        m.Name,
			Kind:        classifyKind(obj.Type()),
		}
	}

	return resolved, errs
}
