package client

import (
    "context"
    "flag"
    "os"
    "time"

    "gopkg.in/yaml.v2"
)

const Flag = "flag"

var (
    _ Client = (*flagClient)(nil)
)

func init() {
    _ = RegisterClient(Flag, newFlagClient)
}

type flagClient struct {
    flags   map[string]interface{}
    flagSet *flag.FlagSet
}

func (c *flagClient) Close() error { return nil }
func (c *flagClient) Pull(ctx context.Context, path string) (data []byte, err error) {
    if !c.flagSet.Parsed() {
        _ = c.flagSet.Parse(os.Args[1:])
    }
    ps := map[string]interface{}{}
    c.flagSet.Visit(func(f *flag.Flag) {
        if v, ok := c.flags[f.Name]; ok && v != nil {
            key := f.Name
            switch v.(type) {
            case *string:
                ps[key] = *(v.(*string))
            case *int:
                ps[key] = *(v.(*int))
            case *uint:
                ps[key] = *(v.(*int))
            case *float64:
                ps[key] = *(v.(*float64))
            case *time.Duration:
                ps[key] = *(v.(*time.Duration))
            case *int64:
                ps[key] = *(v.(*int64))
            case *uint64:
                ps[key] = *(v.(*uint64))
            case *bool:
                ps[key] = *(v.(*bool))
            }
        }
    })
    data, err = yaml.Marshal(ps)
    return
}
func (c *flagClient) Push(ctx context.Context, path string, data []byte) error { return nil }
func (c *flagClient) Watch(path string, cb ChangedCallback) error {
    return nil
}

func newFlagClient(cfg *Config) (Client, error) {
    if cfg.FlagSet == nil {
        cfg.FlagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    }
    if cfg.Flags == nil {
        cfg.Flags = make(map[string]interface{})
    }
    return &flagClient{flags: cfg.Flags, flagSet: cfg.FlagSet}, nil
}
