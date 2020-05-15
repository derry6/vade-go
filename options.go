package vade

import (
    "github.com/derry6/vade-go/pkg/expander"
    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source"
    "github.com/derry6/vade-go/source/client"
)

const (
    DefaultFilePriority   = 0
    DefaultEnvPriority    = 1
    DefaultFlagPriority   = 3
    DefaultRemotePriority = 9
)

var (
    defaultFileOpts = []source.Option {
        source.WithPriority(DefaultFilePriority),
    }
    defaultEnvOpts = []source.Option {
        source.WithPriority(DefaultEnvPriority),
    }
    defaultFlagOpts = []source.Option {
        source.WithPriority(DefaultFlagPriority),
    }
    defaultRemoteOpts = []source.Option {
        source.WithPriority(DefaultRemotePriority),
    }
)

type Option func(opts *options)

type remoteConfig struct {
    name   string
    config *client.Config
    opts   []source.Option
}

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

    // remote Source
    remotes map[string]remoteConfig
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

func WithFileSource(requires, optionals []string, sOpts ...source.Option) Option {
    return func(opts *options) {
        opts.withFile = true
        opts.requireds = requires
        opts.optionals = optionals
        opts.fileOpts = append(defaultFileOpts, sOpts...)
    }
}

func WithEnvSource(sOpts ...source.Option) Option {
    return func(opts *options) {
        opts.withEnv = true
        opts.envOpts = append(defaultEnvOpts, sOpts...)
    }
}
func WithFlagSource(sOpts ...source.Option) Option {
    return func(opts *options) {
        opts.withFlag = true
        opts.flagOpts = append(defaultFlagOpts, sOpts...)
    }
}

func WithRemoteSource(name string, config *client.Config, sOpts ...source.Option) Option {
    return func(opts *options) {
        opts.remotes[name] = remoteConfig{
            name:   name,
            config: config,
            opts:   append(defaultRemoteOpts, sOpts...),
        }
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
        remotes:  make(map[string]remoteConfig),
    }
    for _, o := range opts {
        o(mOpts)
    }
    return mOpts
}
