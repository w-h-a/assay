package service

func toErrors[T error](errs []T) []error {
	out := make([]error, len(errs))
	for i, e := range errs {
		out[i] = e
	}
	return out
}
