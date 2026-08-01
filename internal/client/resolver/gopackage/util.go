package gopackage

import (
	"context"
	"fmt"
	"go/types"

	"github.com/w-h-a/assay/internal/client/resolver"
	"golang.org/x/tools/go/packages"
)

func loadPackage(ctx context.Context, path string) (*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("load package %q: %w", path, err)
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("path %q did not name exactly one package (got %d)", path, len(pkgs))
	}

	pkg := pkgs[0]

	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package %q has build errors: %w", path, pkg.Errors[0])
	}

	if pkg.Types == nil {
		return nil, fmt.Errorf("package %q loaded without type information", path)
	}

	return pkg, nil
}

func resolveFunc(fn *types.Func) resolver.ResolvedFunc {
	sig := fn.Signature()

	resolved := resolver.ResolvedFunc{
		PackagePath: fn.Pkg().Path(),
		Name:        fn.Name(),
		Variadic:    sig.Variadic(),
	}

	params := sig.Params()
	for v := range params.Variables() {
		resolved.Params = append(resolved.Params, resolveTypeRef(v.Type()))
	}

	results := sig.Results()
	for v := range results.Variables() {
		resolved.Returns = append(resolved.Returns, resolveTypeRef(v.Type()))
	}

	return resolved
}

func resolveTypeRef(t types.Type) resolver.ResolvedType {
	ref := resolver.ResolvedType{
		Kind: classifyKind(t),
		Expr: types.TypeString(t, func(p *types.Package) string { return p.Name() }),
	}

	if basic, ok := t.(*types.Basic); ok {
		ref.Name = basic.Name()
		return ref
	}

	named, ok := namedOf(t)
	if !ok {
		return ref
	}

	ref.Name = named.Obj().Name()
	ref.Kind = classifyKind(named)

	if pkg := named.Obj().Pkg(); pkg != nil {
		ref.PackagePath = pkg.Path()
	}

	return ref
}

func namedOf(t types.Type) (*types.Named, bool) {
	t = types.Unalias(t)

	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}

	named, ok := t.(*types.Named)

	return named, ok
}

func classifyKind(t types.Type) resolver.Kind {
	switch t.Underlying().(type) {
	case *types.Struct:
		return resolver.KindStruct
	case *types.Interface:
		return resolver.KindInterface
	default:
		return resolver.KindBasic
	}
}
