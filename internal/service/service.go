package service

import (
	"os"

	"github.com/w-h-a/assay/internal/spec/checker"
	"github.com/w-h-a/assay/internal/spec/parser"
)

// Service orchestrates assay operations.
type Service struct{}

func New() *Service {
	return &Service{}
}

// Check reads a spec file, parses it, and type-checks it.
func (s *Service) Check(specPath string) []error {
	source, err := os.ReadFile(specPath)
	if err != nil {
		return []error{err}
	}

	spec, parseErrs := parser.Parse(string(source), specPath)
	if len(parseErrs) > 0 {
		return toErrors(parseErrs)
	}

	_, checkErrs := checker.Check(spec)
	if len(checkErrs) > 0 {
		return toErrors(checkErrs)
	}

	return nil
}
