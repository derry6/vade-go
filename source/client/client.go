package client

import (
    "context"
    "io"

    pkgerrs "github.com/pkg/errors"
)

var (
    clientConstructors = map[string]Constructor{}
)

type ChangedCallback func(data []byte)

// Client client interface of baseSource
type Client interface {
    io.Closer
    Pull(ctx context.Context, path string) (data []byte, err error)
    Push(ctx context.Context, path string, data []byte) error
    Watch(path string, cb ChangedCallback) error
}

type Constructor func(config *Config) (Client, error)

func RegisterClient(name string, constructor Constructor) error {
    clientConstructors[name] = constructor
    return nil
}

func New(name string, config *Config) (Client, error) {
    if config == nil {
        config = DefaultConfig()
    }
    if c, ok := clientConstructors[name]; ok {
        return c(config)
    }
    return nil, pkgerrs.Errorf("client constructor of %q not exists", name)
}
