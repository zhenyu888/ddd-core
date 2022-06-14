package ebus

type Option func(opts *Options)

type Options struct {
	dispatcher Dispatcher
}

func buildOptions(opts ...Option) *Options {
	rlt := &Options{}
	for _, opt := range opts {
		opt(rlt)
	}
	fillDefaults(rlt)
	return rlt
}

func fillDefaults(opts *Options) {
	if opts.dispatcher == nil {
		opts.dispatcher = NewImmediateDispatcher()
	}
}

func WithDispatcher(dispatcher Dispatcher) Option {
	return func(opts *Options) {
		opts.dispatcher = dispatcher
	}
}
