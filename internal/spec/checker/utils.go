package checker

import "github.com/w-h-a/assay/internal/spec/ast"

func isNumeric(typeName string) bool {
	return typeName == "int" || typeName == "uint" || typeName == "float"
}

func literalType(kind ast.LiteralKind) string {
	switch kind {
	case ast.LiteralInt:
		return "int"
	case ast.LiteralFloat:
		return "float"
	case ast.LiteralString:
		return "string"
	case ast.LiteralBool:
		return "bool"
	case ast.LiteralBytes:
		return "bytes"
	default:
		return ""
	}
}
