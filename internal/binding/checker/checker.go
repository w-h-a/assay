package checker

import (
	"fmt"

	"github.com/w-h-a/assay/internal/binding/ast"
	specast "github.com/w-h-a/assay/internal/spec/ast"
)

// Error represents a binding cross-reference error with source position.
type Error struct {
	Message string
	Pos     ast.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Check validates a parsed binding against a validated spec. It verifies
// that the binding's name matches the spec's name and that every type and
// func mapping references a declaration that exists in the spec. The
// validated binding is always returned, even when errors are present, so
// callers can report multiple problems.
//
// Go-side resolution (qualifier.Name against the target package) is the
// resolver's concern, not this pass.
func Check(binding *ast.BindingDecl, spec *specast.ValidatedSpec) (*ast.ValidatedBinding, []Error) {
	c := &checker{
		types: map[string]struct{}{},
		funcs: map[string]struct{}{},
	}

	c.indexSpec(spec.Spec)
	c.checkBindingName(binding, spec.Spec)
	c.checkTypeMappings(binding.TypeMappings)
	c.checkFuncMappings(binding.FuncMappings)

	return &ast.ValidatedBinding{Binding: binding}, c.errors
}

// checker collects the spec's declared type and func names, then walks the
// binding's mappings looking for references that don't resolve. Errors are
// collected rather than halting.
type checker struct {
	types  map[string]struct{}
	funcs  map[string]struct{}
	errors []Error
}

// indexSpec registers the names of every TypeDecl and FuncDecl in the spec.
// Empty names from parser recovery are skipped — the spec parser already
// reported the missing identifier.
func (c *checker) indexSpec(spec *specast.SpecDecl) {
	for _, d := range spec.Declarations {
		switch d := d.(type) {
		case *specast.TypeDecl:
			if d.Name == "" {
				continue
			}
			c.types[d.Name] = struct{}{}
		case *specast.FuncDecl:
			if d.Name == "" {
				continue
			}
			c.funcs[d.Name] = struct{}{}
		}
	}
}

// checkBindingName reports an error when the binding name does not match
// the spec name. Empty names from parser recovery are skipped.
func (c *checker) checkBindingName(binding *ast.BindingDecl, spec *specast.SpecDecl) {
	if binding.Name == "" || spec.Name == "" {
		return
	}
	if binding.Name != spec.Name {
		c.addError(binding.Pos, "binding name %q does not match spec name %q", binding.Name, spec.Name)
	}
}

func (c *checker) checkTypeMappings(mappings []ast.TypeMapping) {
	for _, m := range mappings {
		if m.SpecName == "" {
			continue
		}
		if _, ok := c.types[m.SpecName]; !ok {
			c.addError(m.Pos, "type mapping references undeclared spec type %q", m.SpecName)
		}
	}
}

func (c *checker) checkFuncMappings(mappings []ast.FuncMapping) {
	for _, m := range mappings {
		if m.SpecName == "" {
			continue
		}
		if _, ok := c.funcs[m.SpecName]; !ok {
			c.addError(m.Pos, "func mapping references undeclared spec func %q", m.SpecName)
		}
	}
}

func (c *checker) addError(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     pos,
	})
}
