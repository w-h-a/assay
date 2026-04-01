package verdictstore

import "context"

// Option is a functional option for VerdictStore configuration
type Option func(*Options)

// Options configures a VerdictStore implementation.
type Options struct {
	BaseLocation string
	Context      context.Context
}

// WithBaseLocation sets the base location for verdict storage.
func WithBaseLocation(loc string) Option {
	return func(o *Options) {
		o.BaseLocation = loc
	}
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
