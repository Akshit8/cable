package cable

import "context"

// Option is a field in Options
type Option interface {
	Change(*Options)
}

// OptionSetter is a function that updates the options object with new option.
type OptionSetter func(*Options)

func (s OptionSetter) Change(options *Options) {
	s(options)
}

func WithContext(ctx context.Context) Option {
	return OptionSetter(func(options *Options) {
		options.ctx = ctx
	})
}

func WithLogger(logger Logger) Option {
	return OptionSetter(func(options *Options) {
		options.logger = logger
	})
}

type Options struct {
	ctx    context.Context
	logger Logger
}

func newOptions(option ...Option) *Options {
	defaultOptions := &Options{
		ctx:    context.Background(),
		logger: NewLogger(),
	}

	for _, opt := range option {
		opt.Change(defaultOptions)
	}

	return defaultOptions
}
