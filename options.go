package vade

import (
    "github.com/derry6/vade-go/pkg/expander"
    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source"
)

const (
    DefaultFilePriority   = 0
    DefaultEnvPriority    = 1
    DefaultFlagPriority   = 3
    DefaultRemotePriority = 9
)

type Option func(opts *options)

type options struct {
    // fileSource options
    withFile  bool
    requireds []string
    optionals []string
    fileOpts  []source.Option
    // envSource options
    withEnv bool
    envOpts []source.Option
    // flagSource options
    withFlag bool
    flagOpts []source.Option
    // logger
    logger log.Logger
    // expansion
    epOpts     []expander.Option
    epDisabled bool
}

func WithLogger(logger log.Logger) Option {
    return func(opts *options) {
        opts.logger = logger
    }
}

func WithFileSource(requires, optionals []string) Option {
    return func(opts *options) {
        opts.withFile = true
        opts.requireds = requires
        opts.optionals = optionals
    }
}
func WithFileSourcePrefix(prefix string) Option {
    return func(opts *options) {
        opts.fileOpts = append(opts.fileOpts, source.WithPrefix(prefix))
    }
}
func WithFileSourcePriority(pri int) Option {
    return func(opts *options) {
        opts.fileOpts = append(opts.fileOpts, source.WithPriority(pri))
    }
}

func WithEnvSource(sOpts ...source.Option) Option {
    return func(opts *options) {
        opts.withEnv = true
        opts.envOpts = sOpts
    }
}
func WithEnvSourcePrefix(prefix string) Option {
    return func(opts *options) {
        opts.envOpts = append(opts.envOpts, source.WithPrefix(prefix))
    }
}
func WithEnvSourcePriority(pri int) Option {
    return func(opts *options) {
        opts.envOpts = append(opts.envOpts, source.WithPriority(pri))
    }
}

func WithFlagSource() Option {
    return func(opts *options) {
        opts.withFlag = true
    }
}
func WithFlagSourcePrefix(prefix string) Option {
    return func(opts *options) {
        opts.flagOpts = append(opts.flagOpts, source.WithPrefix(prefix))
    }
}
func WithFlagSourcePriority(pri int) Option {
    return func(opts *options) {
        opts.flagOpts = append(opts.flagOpts, source.WithPriority(pri))
    }
}

// WithExpandDisabled 禁止变量替换
func WithExpandDisabled() Option {
    return func(opts *options) {
        opts.epDisabled = true
    }
}

// WithExpansion 指定需要变量替换的模板
func WithExpansion(pre, post string, cb expander.Handler) Option {
    return func(opts *options) {
        opts.epOpts = append(opts.epOpts, expander.WithExpansion(pre, post, cb))
    }
}

func newOptions(opts ...Option) *options {
    mOpts := &options{
        fileOpts: []source.Option{source.WithPriority(DefaultFilePriority)},
        envOpts:  []source.Option{source.WithPriority(DefaultEnvPriority)},
        flagOpts: []source.Option{source.WithPriority(DefaultFlagPriority)},
    }
    for _, o := range opts {
        o(mOpts)
    }
    return mOpts
}
