package vade

import (
    "github.com/derry6/vade-go/pkg/expander"
    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source"
    "github.com/derry6/vade-go/source/client"
)

func (mgr *manager) initFlagSource(opts ...source.Option) error {
    cfg := client.DefaultConfig()
    cfg.Flags = _bindFlags
    cfg.FlagSet = _flagSet
    c, err := client.New(client.Flag, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.Flag, c, opts...)
    _ = s.AddPath("default", source.WithPathRequired())
    return mgr.AddSource(s)
}

func (mgr *manager) initEnvSource(opts ...source.Option) error {
    cfg := client.DefaultConfig()
    c, err := client.New(client.Env, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.Env, c, opts...)
    _ = s.AddPath("default", source.WithPathRequired())
    return mgr.AddSource(s)
}

func (mgr *manager) initFileSource(requires, optionals []string, opts ...source.Option) error {
    cfg := client.DefaultConfig()
    c, err := client.New(client.File, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.File, c, opts...)
    for _, require := range requires {
        if err = filesInDir(require, func(f string) error {
            return s.AddPath(f, source.WithPathRequired())
        }); err != nil {
            return err
        }
    }
    for _, optional := range optionals {
        if err = filesInDir(optional, func(f string) error {
            return s.AddPath(f)
        }); err != nil {
            return err
        }
    }
    return mgr.AddSource(s)
}

func (mgr *manager) initRemoteSource(remotes map[string]remoteConfig) error {
    for _, remote := range remotes {
        cli, err := client.New(remote.name, remote.config)
        if err != nil {
            return err
        }
        s := source.New(remote.name, cli, remote.opts...)
        if err = mgr.AddSource(s); err != nil {
            return err
        }
    }
    return nil
}

func (mgr *manager) init(vOpts *options) (err error) {
    if vOpts.logger != nil {
        SetLogger(vOpts.logger)
    }
    mgr.expander = expander.New(mgr.unsafeGet, vOpts.epOpts...)
    mgr.expandDisabled = vOpts.epDisabled
    if vOpts.withFile {
        if err = mgr.initFileSource(vOpts.requireds, vOpts.optionals, vOpts.fileOpts...); err != nil {
            return err
        }
    }
    if vOpts.withEnv {
        if err = mgr.initEnvSource(vOpts.envOpts...); err != nil {
            return err
        }
    }
    if vOpts.withFlag {
        if err = mgr.initFlagSource(vOpts.flagOpts...); err != nil {
            return err
        }
    }
    return mgr.initRemoteSource(vOpts.remotes)
}

func SetLogger(logger log.Logger) {
    log.Use(logger)
}

func Init(opts ...Option) (err error) {
    _mgr, err = newManager(opts...)
    return err
}
