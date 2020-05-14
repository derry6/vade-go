package vade

import (
    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source"
    "github.com/derry6/vade-go/source/client"
)

func initFlagSource(mgr Manager, opts *options) error  {
    cfg := client.DefaultConfig()
    cfg.Flags = _bindFlags
    cfg.FlagSet = _flagSet
    c, err := client.New(client.Flag, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.Flag, c, opts.flagOpts...)
    _ = s.AddPath("default", source.WithPathRequired())
    return mgr.AddSource(s)
}


func initEnvSource(mgr Manager, opts *options) error {
    cfg := client.DefaultConfig()
    c, err := client.New(client.Env, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.Env, c, opts.envOpts...)
    _ = s.AddPath("default", source.WithPathRequired())
    return mgr.AddSource(s)
}


func initFileSource(mgr Manager, opts *options) error {
    cfg := client.DefaultConfig()
    c, err := client.New(client.File, cfg)
    if err != nil {
        return err
    }
    s := source.New(client.File, c, opts.fileOpts...)
    for _, require := range opts.requireds {
        if err = filesInDir(require, func(f string) error {
            return s.AddPath(f, source.WithPathRequired())
        }); err != nil {
            return err
        }
    }
    for _, optional := range opts.optionals {
        if err = filesInDir(optional, func(f string) error {
            return s.AddPath(f)
        }); err != nil {
            return err
        }
    }
    return mgr.AddSource(s)
}


func Init(opts ...Option) (err error) {
    vOpts := newOptions(opts...)
    log.Use(vOpts.logger)
    switch m := mgr.(type) {
    case *manager:
        m.setOptions(vOpts)
    }
    if vOpts.withFile {
        if err = initFileSource(mgr, vOpts); err != nil {
            return err
        }
    }
    if vOpts.withEnv {
        if err = initEnvSource(mgr, vOpts); err != nil {
            return err
        }
    }
    if vOpts.withFlag {
       if err = initFlagSource(mgr, vOpts); err != nil {
           return err
       }
    }
    return nil
}
