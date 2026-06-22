package resolver

import "context"

// Option is a functional option for Resolver configuration
type Option func(*Options)

// Options configures a Resolver implementation.
type Options struct {
	Context context.Context
}

func NewOptions(opts ...Option) Options {
	options := Options{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}
