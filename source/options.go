package source

import "github.com/derry6/vade-go/source/parser"

type options struct {
    priority    int
    prefix      string
    withDeleted bool
}

type Option func(opts *options)

func WithPrefix(prefix string) Option {
    return func(opts *options) {
        opts.prefix = prefix
    }
}

func WithPriority(priority int) Option {
    return func(opts *options) {
        opts.priority = priority
    }
}

func WithDeleted() Option {
    return func(opts *options) {
        opts.withDeleted = true
    }
}

func newOptions(opts ...Option) *options {
    sOpts := &options{prefix: ""}
    for _, o := range opts {
        o(sOpts)
    }
    return sOpts
}

type pathOptions struct {
    priority      int
    required      bool
    parser        parser.Parser
    watchDisabled bool
}

type PathOption func(opts *pathOptions)

func WithPathPriority(pri int) PathOption {
    return func(opts *pathOptions) {
        opts.priority = pri
    }
}

func WithPathRequired() PathOption {
    return func(opts *pathOptions) {
        opts.required = true
    }
}

func WithPathParser(parser parser.Parser) PathOption {
    return func(opts *pathOptions) {
        opts.parser = parser
    }
}

func WithPathWatchDisabled() PathOption {
    return func(opts *pathOptions) {
        opts.watchDisabled = true
    }
}

func newPathOptions(opts ...PathOption) *pathOptions {
    nsOpts := &pathOptions{
        priority:      0,
        required:      false,
        parser:        nil,
        watchDisabled: false,
    }
    for _, optFn := range opts {
        optFn(nsOpts)
    }
    return nsOpts
}
