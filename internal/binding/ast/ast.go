package ast

import "fmt"

// BindingDecl is the root node representing the contents of a .bind file.
type BindingDecl struct {
	Name         string
	Target       string
	PackagePath  string
	TypeMappings []TypeMapping
	FuncMappings []FuncMapping
	Pos          Position
}

// TypeMapping connects a spec type name to a qualifier.name in the target.
// The resolver validates the reference against the target package.
type TypeMapping struct {
	SpecName  string
	Qualifier string
	Name      string
	Pos       Position
}

// FuncMapping connects a spec function name to a qualifier.name in the target.
// Whether Qualifier names a package alias (pkg.Func) or a receiver type
// (Type.Method) is determined by the resolver against the spec's first
// parameter and the target package — not at parse time.
type FuncMapping struct {
	SpecName  string
	Qualifier string
	Name      string
	Pos       Position
}

// Position tracks a location in source code.
type Position struct {
	File   string
	Line   int
	Column int
}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}
