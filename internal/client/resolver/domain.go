package resolver

// Kind classifies a resolved target type by its shape, so a backend can
// generate code without re-inspecting the target package.
type Kind string

const (
	KindBasic     Kind = "basic"
	KindStruct    Kind = "struct"
	KindInterface Kind = "interface"
)

// ResolvedType is a target-language type: where it lives, its name, its
// kind, and how the type is written out in the target language.
type ResolvedType struct {
	PackagePath string
	Name        string
	Kind        Kind
	Expr        string
}

// ResolvedFunc is a target-language function or method signature.
type ResolvedFunc struct {
	PackagePath string
	Name        string
	Receiver    *ResolvedType
	Variadic    bool
	Params      []ResolvedType
	Returns     []ResolvedType
}

// ResolvedBinding is the resolver's output.
type ResolvedBinding struct {
	Funcs map[string]ResolvedFunc
	Types map[string]ResolvedType
}
