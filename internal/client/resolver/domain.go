package resolver

// Kind classifies a resolved target type by its shape, so a backend can
// generate code without re-inspecting the target package.
type Kind string

const (
	KindBasic     Kind = "basic"
	KindStruct    Kind = "struct"
	KindInterface Kind = "interface"
)

// ResolvedType is a target-language type: where it lives, its name, and its
// shape.
type ResolvedType struct {
	PackagePath string
	Name        string
	Kind        Kind
}

// ResolvedFunc is a target-language function or method signature.
type ResolvedFunc struct {
	PackagePath string
	Name        string
	Receiver    *ResolvedType
	Params      []ResolvedType
	Returns     []ResolvedType
}

// ResolvedBinding is the resolver's output.
type ResolvedBinding struct {
	Funcs map[string]ResolvedFunc
	Types map[string]ResolvedType
}
