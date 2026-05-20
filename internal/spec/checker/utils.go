package checker

import (
	"strings"

	"github.com/w-h-a/assay/internal/spec/ast"
)

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

func includesError(typeName string) bool {
	if typeName == "error" {
		return true
	}
	if len(typeName) < 3 || typeName[0] != '(' || typeName[len(typeName)-1] != ')' {
		return false
	}
	inner := typeName[1 : len(typeName)-1]
	depth := 0
	start := 0
	for i := 0; i < len(inner); i++ {
		switch inner[i] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ',':
			if depth == 0 {
				if strings.TrimSpace(inner[start:i]) == "error" {
					return true
				}
				start = i + 1
			}
		}
	}
	return strings.TrimSpace(inner[start:]) == "error"
}

func parseTupleElements(typeName string) []string {
	if len(typeName) < 3 || typeName[0] != '(' || typeName[len(typeName)-1] != ')' {
		return nil
	}
	inner := typeName[1 : len(typeName)-1]
	var elements []string
	depth := 0
	start := 0
	for i := 0; i < len(inner); i++ {
		switch inner[i] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ',':
			if depth == 0 {
				elements = append(elements, strings.TrimSpace(inner[start:i]))
				start = i + 1
			}
		}
	}
	elements = append(elements, strings.TrimSpace(inner[start:]))
	return elements
}
