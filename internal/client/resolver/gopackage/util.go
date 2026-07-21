package gopackage

import (
	"context"
	"fmt"
	"go/types"

	"github.com/w-h-a/assay/internal/client/resolver"
	"golang.org/x/tools/go/packages"
)

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
